// Package model はドメインエンティティ(GORMモデルを兼ねる)が共通で埋め込む
// 構造体を定義する(architecture.md「domain＝DBモデル・BaseModel・楽観ロック」)。
package model

import "time"

// BaseModel はID採番・作成/更新日時・楽観ロック用バージョンを表す。
// 新規ドメインエンティティはこれを埋め込むことでこれらのカラムを持つ。
//
// これらのフィールドへの代入はリポジトリや変換関数の中で行わない。
//   - ID / Version: internal/database/mysql の BaseModelPlugin (GORM Callbacks) が
//     Create時に採番する。
//   - CreatedAt / UpdatedAt: GORMの標準機能(フィールド名によるautoCreateTime/
//     autoUpdateTimeの自動認識)が設定する。
//   - Version: 更新時はBaseModelPluginがインクリメントし、
//     `WHERE version = <更新前の値>` を付与して楽観ロックを行う。
type BaseModel struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Version   int       `gorm:"column:version" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
}
