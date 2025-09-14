package debug

import (
	"bytes"
	"testing"

	"github.com/goforj/godump"
	"github.com/stretchr/testify/assert"
)

type info struct {
	Name string
	embed
}

type embed struct {
	Foo string
	Bar struct {
		Baz string
	}
}

func captureOutput(fn func()) string {
	var wirer bytes.Buffer

	originDumper, originWriter := dumper, writer
	defer func() {
		dumper, writer = originDumper, originWriter
	}()

	dumper = godump.NewDumper(godump.WithWriter(&wirer))
	writer = &wirer

	fn()

	return wirer.String()
}

func TestDd(t *testing.T) {
	origin := osExit
	osExit = func(int) {}
	t.Cleanup(func() {
		osExit = origin
	})

	output := captureOutput(func() {
		DD(struct{ ID int }{ID: 123})
	})
	assert.Equal(t, "\x1b[90m<#dump // dump.go:19"+
		"\x1b[0m\n\x1b[90m#struct { ID int }\x1b[0m {\n  "+
		"\x1b[33m+\x1b[0mID => \x1b[38;5;38m123\x1b[0m\n}\n",
		output)
}

func TestDump(t *testing.T) {
	data := info{
		Name: "test",
		embed: embed{
			Foo: "foo",
			Bar: struct {
				Baz string
			}{
				Baz: "baz",
			},
		},
	}

	t.Run("dump", func(t *testing.T) {
		output := captureOutput(func() {
			Dump(data)
		})
		assert.Equal(t, "\x1b[90m<#dump // dump.go:25"+
			"\x1b[0m\n\x1b[90m#debug.info\x1b[0m {\n  "+
			"\x1b[33m+\x1b[0mName    => \x1b[33m\"\x1b[0m\x1b[1;38;5;113mtest"+
			"\x1b[0m\x1b[33m\"\x1b[0m\n  \x1b[33m-\x1b[0membed   => \x1b[90m#debug.embed"+
			"\x1b[0m {\n    \x1b[33m+\x1b[0mFoo   => \x1b[33m\"\x1b[0m\x1b[1;38;5;113mfoo"+
			"\x1b[0m\x1b[33m\"\x1b[0m\n    \x1b[33m+\x1b[0mBar   => \x1b[90m#struct { Baz string }"+
			"\x1b[0m {\n      \x1b[33m+\x1b[0mBaz => \x1b[33m\"\x1b[0m\x1b[1;38;5;113mbaz"+
			"\x1b[0m\x1b[33m\"\x1b[0m\n    }\n  }\n}\n",
			output)
	})

	t.Run("dump with HTML", func(t *testing.T) {
		output := captureOutput(func() {
			DumpHTML(data)
		})
		assert.Equal(t, "<div style='background-color:black;'><pre style=\"background-color:black; color:white; padding:5px; border-radius: 5px\">\n"+
			"<span style=\"color:#999\"><#dump // dump.go:30</span>\n<span style=\"color:#999\">#debug.info</span> {\n  "+
			"<span style=\"color:#ffb400\">+</span>Name    => <span style=\"color:#ffb400\">\"</span><span style=\"color:#80ff80\">test"+
			"</span><span style=\"color:#ffb400\">\"</span>\n  <span style=\"color:#ffb400\">-</span>embed   => <span style=\"color:#999\">#debug.embed</span> {\n"+
			"    <span style=\"color:#ffb400\">+</span>Foo   => <span style=\"color:#ffb400\">\"</span><span style=\"color:#80ff80\">foo</span>"+
			"<span style=\"color:#ffb400\">\"</span>\n    <span style=\"color:#ffb400\">+</span>Bar   => <span style=\"color:#999\">#struct { Baz string }</span> {\n"+
			"      <span style=\"color:#ffb400\">+</span>Baz => <span style=\"color:#ffb400\">\"</span><span style=\"color:#80ff80\">baz</span>"+
			"<span style=\"color:#ffb400\">\"</span>\n    }\n  }\n}\n</pre></div>\n", output)
	})

	t.Run("dump with JSON", func(t *testing.T) {
		output := captureOutput(func() {
			DumpJSON(data)
		})
		assert.Equal(t, "{\n  \"Name\": \"test\",\n  \"Foo\": \"foo\",\n  \"Bar\": {\n    \"Baz\": \"baz\"\n  }\n}\n", output)
	})

}

func TestFDump(t *testing.T) {
	data := map[string]string{"key": "value"}

	t.Run("dump to writer", func(t *testing.T) {
		var buf bytes.Buffer
		FDump(&buf, data)
		assert.Equal(t, "\x1b[90m<#dump // dump_test.go:111\x1b[0m\n{\n   "+
			"\x1b[38;5;170mkey\x1b[0m => \x1b[33m\"\x1b[0m\x1b[1;38;5;113mvalue"+
			"\x1b[0m\x1b[33m\"\x1b[0m\n}\n",
			buf.String())

	})

	t.Run("dump HTML to writer", func(t *testing.T) {
		var buf bytes.Buffer
		FDumpHTML(&buf, data)
		assert.Equal(t, "<div style='background-color:black;'><pre style=\"background-color:black; color:white; padding:5px; border-radius: 5px\">\n"+
			"<span style=\"color:#999\"><#dump // dump_test.go:121</span>\n{\n   <span style=\"color:#d087d0\">key</span> => <span style=\"color:#ffb400\">\"</span>"+
			"<span style=\"color:#80ff80\">value</span><span style=\"color:#ffb400\">\"</span>\n}\n</pre></div>\n",
			buf.String())

	})

	t.Run("dump JSON to writer", func(t *testing.T) {
		var buf bytes.Buffer
		FDumpJSON(&buf, data)
		assert.Equal(t, "{\n  \"key\": \"value\"\n}\n", buf.String())
	})

}

func TestSDump(t *testing.T) {
	data := []string{"one", "two"}

	t.Run("dump as string", func(t *testing.T) {
		output := SDump(data)
		assert.Equal(t, "\x1b[90m<#dump // dump_test.go:141\x1b[0m\n[\n  "+
			"\x1b[38;5;38m0\x1b[0m => \x1b[33m\"\x1b[0m\x1b[1;38;5;113mone"+
			"\x1b[0m\x1b[33m\"\x1b[0m\n  \x1b[38;5;38m1\x1b[0m => \x1b[33m\""+
			"\x1b[0m\x1b[1;38;5;113mtwo\x1b[0m\x1b[33m\"\x1b[0m\n]\n",
			output)
	})

	t.Run("dump HTML as string", func(t *testing.T) {
		output := SDumpHTML(data)
		assert.Equal(t, "<div style='background-color:black;'><pre style=\"background-color:black; color:white; padding:5px; border-radius: 5px\">\n"+
			"<span style=\"color:#999\"><#dump // dump_test.go:150</span>\n[\n  <span style=\"color:#40c0ff\">0</span> => <span style=\"color:#ffb400\">\"</span>"+
			"<span style=\"color:#80ff80\">one</span><span style=\"color:#ffb400\">\"</span>\n  <span style=\"color:#40c0ff\">1</span> => "+
			"<span style=\"color:#ffb400\">\"</span><span style=\"color:#80ff80\">two</span><span style=\"color:#ffb400\">\"</span>\n]\n</pre></div>",
			output)
	})

	t.Run("dump JSON as string", func(t *testing.T) {
		output := SDumpJSON(data)
		assert.Equal(t, "[\n  \"one\",\n  \"two\"\n]", output)
	})

}
