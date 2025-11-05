package plugins

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/internal/telemetry"
)

func LogToRoute(r *http.Request, txt interface{}) {
	items := strings.Split(r.URL.String(), "/")
	if len(items) > 3 {
		ns := items[2]
		path := filepath.Join("/", strings.Join(items[3:], "/"))
		telemetry.LogRoute(telemetry.LogLevelInfo, ns, path, fmt.Sprintf("%v", txt))
	} else {
		slog.Error("can not parse route url in js log function")
	}
}
