package internel

import "github.com/samber/lo"

type TextMap struct {
	data map[string]string
}

func NewTextMap() *TextMap {
	return &TextMap{
		data: make(map[string]string),
	}
}

// Get implements TextMapCarrier.
func (t *TextMap) Get(key string) string {
	return t.data[key]
}

// Keys implements TextMapCarrier.
func (t *TextMap) Keys() []string {
	return lo.Keys(t.data)
}

// Set implements TextMapCarrier.
func (t *TextMap) Set(key string, value string) {
	t.data[key] = value
}
