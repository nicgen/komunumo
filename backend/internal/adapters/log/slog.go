package log

import (
	"io"
	"log/slog"
	"strings"
)

// sensitiveKeys are redacted in the structured log output regardless of where
// they appear in the attribute tree. Matched case-insensitively as suffixes
// so that "user.email" or "request.body.password" are caught too.
var sensitiveKeys = []string{
	"password",
	"password_hash",
	"passwordhash",
	"token",
	"raw_token",
	"rawtoken",
	"token_hash",
	"tokenhash",
	"email",
	"authorization",
	"cookie",
	"set-cookie",
}

// New returns a JSON slog logger that redacts sensitive attribute values.
func New(w io.Writer, level slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redact,
	})
	return slog.New(h)
}

func redact(_ []string, a slog.Attr) slog.Attr {
	k := strings.ToLower(a.Key)
	for _, s := range sensitiveKeys {
		if k == s || strings.HasSuffix(k, "."+s) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}
	return a
}
