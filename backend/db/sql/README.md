# db/sql

`docker compose up` 時に MySQL コンテナの `docker-entrypoint-initdb.d` としてマウントされるディレクトリ（`compose.yml` 参照）。

**スキーマ定義はここには置かない。** テーブル定義の正は `db/migrations/`（`cmd/migrate` / `sql-migrate` で適用する）。このディレクトリは、DB初回起動時に一度だけ実行しておきたい処理（追加ユーザーの権限付与など）が必要になった場合にのみ `.sql` を追加する。
