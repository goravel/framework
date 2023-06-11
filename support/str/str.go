package str

import (
	"bytes"
	"crypto/rand"
	"strconv"
	"strings"
	"unicode"
)

func Random(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	letters := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, v := range b {
		b[i] = letters[v%byte(len(letters))]
	}

	return string(b)
}

func Case2Camel(name string) string {
	names := strings.Split(name, "_")

	var newName string
	for _, item := range names {
		buffer := NewBuffer()
		for i, r := range item {
			if i == 0 {
				buffer.Append(unicode.ToUpper(r))
			} else {
				buffer.Append(r)
			}
		}

		newName += buffer.String()
	}

	return newName
}

func Camel2Case(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}

	return buffer.String()
}

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i any) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Write(val)
	case rune:
		b.WriteRune(val)
	}
	return b
}

func (b *Buffer) append(s string) *Buffer {
	b.WriteString(s)

	return b
}
