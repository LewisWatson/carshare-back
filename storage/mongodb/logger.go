package mongodb

import (
	"os"

	logging "github.com/op/go-logging"
)

var (
	log    = logging.MustGetLogger("resource")
	format = logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999} %{shortpkg} %{longfunc} %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
)

func init() {
	logging.SetBackend(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format))
}
