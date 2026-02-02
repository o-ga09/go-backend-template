package constant

import (
	"log/slog"

	"cloud.google.com/go/logging"
)

// cloud logging の Log level 定義
var (
	Severitydefault = slog.Level(logging.Default)
	SeverityInfo    = slog.Level(logging.Info)
	SeverityWarn    = slog.Level(logging.Warning)
	SeverityError   = slog.Level(logging.Error)
	SeverityNotice  = slog.Level(logging.Notice)
)

// TODO: サービス名を設定する
const SERVICE_NAME = ""
