// +build cgo
// +build !appengine

package metrics

import (
	base "runtime"
)

func GetNumCgoCall() int64 {
	return base.NumCgoCall()
}
