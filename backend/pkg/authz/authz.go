// Package authz はリソースの所有者本人かどうかを判定する、ドメインに依存しない
// 純粋関数を提供する。
//
// architecture.md は「あるドメインが別ドメインの状態を参照する必要がある場面…
// ドメイン間に新しい矢印を作らない」ことを求めている。所有者チェックは
// user・cart・order など複数の将来ドメイン(#4)が同じ形で必要とするロジックだが、
// ドメイン固有の状態を持たないため、各ドメインが依存し合わずに再利用できるよう
// pkg/ 配下の汎用ユーティリティとして切り出す。
//
// 各ドメインは自身のエンティティに、ここを呼び出す薄いメソッド
// (例: user.User.IsOwnedBy)を用意し、ドメイン層の純粋関数としての体裁を保つ。
package authz

// IsOwner はrequesterID(リクエスト主体のユーザーID)が、
// ownerID(リソースの所有者ユーザーID)本人であるかどうかを判定する。
// どちらかが空文字列の場合は所有者ではないとみなす。
func IsOwner(requesterID, ownerID string) bool {
	return requesterID != "" && ownerID != "" && requesterID == ownerID
}
