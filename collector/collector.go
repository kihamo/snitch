package collector

import (
	"github.com/kihamo/snitch"
)

func init() {
	snitch.DefaultRegisterer.Register(
		NewDebugCollector(),
		NewRuntimeCollector())
}
