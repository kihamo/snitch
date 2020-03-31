package snitch

import (
	"fmt"
	"sort"
	"strings"
)

type Labels []*Label

type Label struct {
	Key   string
	Value string
}

func (l Labels) Len() int {
	return len(l)
}
func (l Labels) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l Labels) Less(i, j int) bool {
	cmp := strings.Compare(l[i].Key, l[j].Key)
	if cmp == 0 {
		return strings.Compare(l[i].Value, l[j].Value) < 0
	}

	return cmp < 0
}

func (l Labels) WithMap(labels map[string]string) Labels {
	ret := make(Labels, 0, len(labels))

	for key, value := range labels {
		ret = append(ret, &Label{
			Key:   key,
			Value: value,
		})
	}

	return l.WithLabels(ret)
}

func (l Labels) WithLabels(labels Labels) Labels {
	return append(l, labels...)
}

func (l Labels) With(labels ...string) Labels {
	if len(labels)%2 != 0 {
		labels = append(labels, "unknown")
	}

	ret := make(Labels, 0, len(labels)/2)

	for i := 1; i < len(labels); i += 2 {
		ret = append(ret, &Label{
			Key:   labels[i-1],
			Value: labels[i],
		})
	}

	return l.WithLabels(ret)
}

func (l Labels) String() string {
	var b strings.Builder

	sort.Sort(l)

	for i, label := range l {
		if i != 0 {
			fmt.Fprint(&b, ",")
		}

		fmt.Fprint(&b, label.String())
	}

	return b.String()
}

func (l Labels) Map() map[string]string {
	lvs := make(map[string]string, len(l))

	for _, label := range l {
		lvs[label.Key] = label.Value
	}

	return lvs
}

func (l *Label) String() string {
	return l.Key + "=" + l.Value
}
