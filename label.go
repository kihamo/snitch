package snitch

import (
	"math"
	"strings"
)

type Labels []*Label

type Label struct {
	Key   string
	Value string
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

	return ret
}

func (l Labels) String() string {
	lvs := make([]string, len(l))

	for i, label := range l {
		lvs[i] = label.String()
	}

	return strings.Join(lvs, ",")
}

func (l *Label) String() string {
	return l.Key + "=" + l.Value
}
