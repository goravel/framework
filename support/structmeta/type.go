package structmeta

import (
	"reflect"
	"strings"
)

type FieldMetadata struct {
	Name        string
	Type        string
	Kind        reflect.Kind
	ReflectType reflect.Type
	Anonymous   bool
	Tag         *TagMetadata
	IndexPath   []int
}

type MethodMetadata struct {
	Name         string
	Receiver     string
	Parameters   []string
	Returns      []string
	ReflectValue reflect.Value
}

type StructMetadata struct {
	Name         string
	Fields       []FieldMetadata
	Methods      []MethodMetadata
	ReflectValue reflect.Value
}

type TagItem struct {
	Key        string
	Value      string
	ValueParts []string
}

type TagMetadata struct {
	Raw   reflect.StructTag
	Items []TagItem
	index map[string]int
}

func NewTagMetadata(raw reflect.StructTag) *TagMetadata {
	t := &TagMetadata{
		Raw:   raw,
		index: make(map[string]int),
	}
	tagStr := string(raw)

	for tagStr != "" {
		tagStr = strings.TrimLeft(tagStr, " ")
		if tagStr == "" {
			break
		}
		i := strings.Index(tagStr, ":")
		if i < 0 {
			break
		}
		key := tagStr[:i]
		tagStr = tagStr[i+1:]
		if len(tagStr) < 2 || tagStr[0] != '"' {
			break
		}
		tagStr = tagStr[1:]
		var valBuf strings.Builder
		escapeNext := false
		var j int
		for j = 0; j < len(tagStr); j++ {
			ch := tagStr[j]
			if escapeNext {
				valBuf.WriteByte(ch)
				escapeNext = false
				continue
			}
			if ch == '\\' {
				escapeNext = true
				continue
			}
			if ch == '"' {
				break
			}
			valBuf.WriteByte(ch)
		}
		rawVal := valBuf.String()
		var parts []string
		for _, part := range strings.Split(rawVal, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}

		idx := len(t.Items)
		t.Items = append(t.Items, TagItem{
			Key:        key,
			Value:      rawVal,
			ValueParts: parts,
		})
		t.index[key] = idx

		if j < len(tagStr) {
			tagStr = tagStr[j+1:]
		} else {
			tagStr = ""
		}
	}
	return t
}

func (r *TagMetadata) Get(key string) string {
	if r == nil {
		return ""
	}
	if idx, ok := r.index[key]; ok {
		return r.Items[idx].Value
	}
	return ""
}

func (r *TagMetadata) GetParts(key string) []string {
	if r == nil {
		return nil
	}
	if idx, ok := r.index[key]; ok {
		return r.Items[idx].ValueParts
	}
	return nil
}

func (r *TagMetadata) HasKey(key string) bool {
	if r == nil {
		return false
	}
	_, exists := r.index[key]
	return exists
}
