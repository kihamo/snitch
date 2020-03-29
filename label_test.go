package snitch

import (
	"testing"
)

func BenchmarkWith(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Labels{}.With(
			"test02", "value02",
			"test03", "value03",
			"test01", "value01",
			"test04", "value04",
		).With(
			"test12", "value12",
			"test13", "value13",
			"test11", "value11",
			"test14", "value14",
		)
	}
}
