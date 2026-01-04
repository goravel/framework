package process

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestOutputWriter_Write_SingleLine(t *testing.T) {
	testKey := "test"

	var receivedKey string
	var receivedType contractsprocess.OutputType
	var receivedLine []byte

	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, testKey, func(typ contractsprocess.OutputType, line []byte, key string) {
		receivedKey = key
		receivedType = typ
		receivedLine = append([]byte{}, line...)
	})

	n, err := writer.Write([]byte("hello\n"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, testKey, receivedKey)
	assert.Equal(t, contractsprocess.OutputTypeStdout, receivedType)
	assert.Equal(t, "hello", string(receivedLine))
}

func TestOutputWriter_Write_MultipleLines(t *testing.T) {
	testKey := "test"

	var keys []string
	var lines []string
	var types []contractsprocess.OutputType

	writer := NewOutputWriter(contractsprocess.OutputTypeStderr, testKey, func(typ contractsprocess.OutputType, line []byte, key string) {
		keys = append(keys, key)
		types = append(types, typ)
		lines = append(lines, string(line))
	})

	n, err := writer.Write([]byte("line1\nline2\nline3\n"))
	assert.NoError(t, err)
	assert.Equal(t, 18, n)
	assert.Equal(t, 3, len(lines))
	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
	assert.Equal(t, []string{testKey, testKey, testKey}, keys)
	assert.Equal(t, []contractsprocess.OutputType{
		contractsprocess.OutputTypeStderr,
		contractsprocess.OutputTypeStderr,
		contractsprocess.OutputTypeStderr,
	}, types)
}

func TestOutputWriter_Write_PartialLines(t *testing.T) {
	var lines []string
	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "tests", func(typ contractsprocess.OutputType, line []byte, _ string) {
		lines = append(lines, string(line))
	})

	// Write partial line
	n, err := writer.Write([]byte("partial"))
	assert.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.Empty(t, lines, "No callback for partial line")

	// Complete the line
	n, err = writer.Write([]byte(" line\n"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, []string{"partial line"}, lines)
}

func TestOutputWriter_Write_BufferHandling(t *testing.T) {
	var lines []string

	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "test", func(typ contractsprocess.OutputType, line []byte, _ string) {
		lines = append(lines, string(line))
	})

	// Write multiple chunks with partial lines
	n, err := writer.Write([]byte("first"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	n, err = writer.Write([]byte(" line\nsecond"))
	assert.NoError(t, err)
	assert.Equal(t, 12, n)

	n, err = writer.Write([]byte(" line\n"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	n, err = writer.Write([]byte("third line without newline"))
	assert.NoError(t, err)
	assert.Equal(t, 26, n)

	n, err = writer.Write([]byte("\nfourth line\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	assert.Equal(t, []string{
		"first line",
		"second line",
		"third line without newline",
		"fourth line",
	}, lines)
}

func TestOutputWriter_Write_EmptyLines(t *testing.T) {
	var lines []string

	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "test", func(typ contractsprocess.OutputType, line []byte, _ string) {
		lines = append(lines, string(line))
	})

	n, err := writer.Write([]byte("\n\n\n"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []string{"", "", ""}, lines)
}

func TestOutputWriter_Write_LineModification(t *testing.T) {
	// Test that modifying the line in the callback doesn't affect future callbacks
	var allLines []string

	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "test", func(typ contractsprocess.OutputType, line []byte, _ string) {
		allLines = append(allLines, string(line))
		// Modify the line - should not affect original buffer
		if len(line) > 0 {
			line[0] = 'X'
		}
	})

	n, err := writer.Write([]byte("line1\nline2\n"))
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, []string{"line1", "line2"}, allLines)
}

func TestOutputWriter_Write_LargeInput(t *testing.T) {
	// Test with a large input that spans multiple internal buffer sizes
	lineCount := 0
	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "test", func(typ contractsprocess.OutputType, line []byte, _ string) {
		lineCount++
	})

	// Create a large buffer with many lines
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString("This is line number ")
		buf.WriteString(string(rune('0' + i%10)))
		buf.WriteString("\n")
	}

	n, err := writer.Write(buf.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, buf.Len(), n)
	assert.Equal(t, 1000, lineCount)
}

func TestOutputWriter_Write_KeyIsPropagated(t *testing.T) {
	var receivedKeys []string

	writer := NewOutputWriter(contractsprocess.OutputTypeStdout, "pipe-123", func(typ contractsprocess.OutputType, line []byte, key string) {
		receivedKeys = append(receivedKeys, key)
	})

	n, err := writer.Write([]byte("first\nsecond\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, []string{"pipe-123", "pipe-123"}, receivedKeys)
}
