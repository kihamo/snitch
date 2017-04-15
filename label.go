package snitch

import (
	"math"
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
	ret := make(Labels, 0, len(l)+len(labels))

	for _, value := range l {
		ret = append(ret, &Label{
			Key:   value.Key,
			Value: value.Value,
		})
	}

	for key, value := range labels {
		ret = append(ret, &Label{
			Key:   key,
			Value: value,
		})
	}

	sort.Sort(ret)
	return ret
}

func (l Labels) WithLabels(labels Labels) Labels {
	ret := make(Labels, 0, len(l)+len(labels))

	for _, value := range l {
		ret = append(ret, &Label{
			Key:   value.Key,
			Value: value.Value,
		})
	}

	for _, label := range labels {
		ret = append(ret, &Label{
			Key:   label.Key,
			Value: label.Value,
		})
	}

	sort.Sort(ret)
	return ret
}

func (l Labels) With(labels ...string) Labels {
	ret := make(Labels, 0, len(l)+int(math.Floor(float64(len(labels)/2)))+1)

	for _, value := range l {
		ret = append(ret, &Label{
			Key:   value.Key,
			Value: value.Value,
		})
	}

	if len(labels) == 0 {
		return ret
	}

	if len(labels)%2 != 0 {
		labels = append(labels, "unknown")
	}

	for i := 1; i < len(labels); i += 2 {
		ret = append(ret, &Label{
			Key:   labels[i-1],
			Value: labels[i],
		})
	}

	sort.Sort(ret)
	return ret
}

func (l Labels) String() string {
	lvs := make([]string, len(l))

	for i, label := range l {
		lvs[i] = label.String()
	}

	return strings.Join(lvs, ",")
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
