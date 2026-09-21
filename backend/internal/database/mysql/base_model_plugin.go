package mysql

import (
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

const baseModelPluginName = "base_model"

// optimisticLockCheckedKey はBefore/AfterのUpdateコールバック間で、
// 楽観ロックのバージョンチェックを適用したかどうかを受け渡すための
// InstanceSetキー。
const optimisticLockCheckedKey = "base_model:optimistic_lock_checked"

// BaseModelPlugin はpkg/model.BaseModelを埋め込んだドメインエンティティ向けの
// GORMプラグイン(architecture.md「domain＝DBモデル・BaseModel・楽観ロック」)。
//
//   - Create時: IDが未設定(ゼロ値)であればUUIDを採番し、Versionを1で初期化する。
//     CreatedAt/UpdatedAtはGORM標準機能(フィールド名によるautoCreateTime/
//     autoUpdateTimeの自動認識)が設定するため、本プラグインでは扱わない。
//   - Update時: Versionをインクリメントし、`WHERE version = <更新前の値>` を
//     付与する(楽観ロック)。更新後にRowsAffectedが0件であれば、他のリクエストが
//     先に更新した(競合)とみなしerrors.ErrOptimisticLockConflictをセットする。
//     呼び出し元(リポジトリ)がこれを判別しerrors.MakeConflictErrorに変換する
//     (`WHERE version = ?`をリポジトリに手書きしない)。
//
// リポジトリや変換関数側でID/Version/タイムスタンプを直接代入しないための仕組み。
type BaseModelPlugin struct{}

// NewBaseModelPlugin はBaseModelPluginを生成する。
func NewBaseModelPlugin() *BaseModelPlugin {
	return &BaseModelPlugin{}
}

// Name はgorm.Pluginインターフェースの実装。
func (p *BaseModelPlugin) Name() string {
	return baseModelPluginName
}

// Initialize はgorm.Pluginインターフェースの実装。Create/UpdateのCallbacksを登録する。
func (p *BaseModelPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gorm:create").Register("base_model:before_create", beforeCreate); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register("base_model:before_update", beforeUpdate); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("base_model:after_update", afterUpdate); err != nil {
		return err
	}
	return nil
}

// バッチINSERT(スライス)の場合、stmt.ReflectValueは[]Xxxそのものになるため、
// 要素ごとにID/Versionを採番する(GORM本体のcallbacks/create.goと同じ
// reflect.Indirect(stmt.ReflectValue.Index(i))パターン)。
func beforeCreate(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil {
		return
	}

	idField := stmt.Schema.LookUpField("ID")
	versionField := stmt.Schema.LookUpField("Version")

	switch stmt.ReflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < stmt.ReflectValue.Len(); i++ {
			setBaseModelFields(db, idField, versionField, reflect.Indirect(stmt.ReflectValue.Index(i)))
		}
	default:
		setBaseModelFields(db, idField, versionField, stmt.ReflectValue)
	}
}

func setBaseModelFields(db *gorm.DB, idField, versionField *schema.Field, value reflect.Value) {
	if idField != nil {
		if _, isZero := idField.ValueOf(db.Statement.Context, value); isZero {
			_ = db.AddError(idField.Set(db.Statement.Context, value, uuid.GenerateID()))
		}
	}
	if versionField != nil {
		if _, isZero := versionField.ValueOf(db.Statement.Context, value); isZero {
			_ = db.AddError(versionField.Set(db.Statement.Context, value, 1))
		}
	}
}

func beforeUpdate(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil {
		return
	}

	versionField := stmt.Schema.LookUpField("Version")
	if versionField == nil {
		return
	}

	currentVersion, isZero := versionField.ValueOf(stmt.Context, stmt.ReflectValue)
	if isZero {
		return
	}
	v, ok := currentVersion.(int)
	if !ok {
		return
	}

	db.Statement.SetColumn("Version", v+1)
	db.Where("version = ?", v)
	db.InstanceSet(optimisticLockCheckedKey, true)
}

func afterUpdate(db *gorm.DB) {
	if _, ok := db.InstanceGet(optimisticLockCheckedKey); !ok {
		return
	}
	if db.Error == nil && db.Statement.RowsAffected == 0 {
		_ = db.AddError(pkgerrors.ErrOptimisticLockConflict)
	}
}
