// +build !cgo appengine

package metrics

func GetNumCgoCall() int64 {
	return 0
}
