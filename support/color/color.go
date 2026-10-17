package color

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	gookitcolor "github.com/gookit/color"
	"github.com/pterm/pterm"

	"github.com/goravel/framework/contracts/support"
	"github.com/goravel/framework/support/env"
)

const (
	FgBlack Color = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
	// FgDefault revert default FG.
	FgDefault Color = 39
)

// Extra foreground color 90 - 97.
const (
	FgDarkGray Color = iota + 90
	FgLightRed
	FgLightGreen
	FgLightYellow
	FgLightBlue
	FgLightMagenta
	FgLightCyan
	FgLightWhite
	// FgGray is an alias of FgDarkGray.
	FgGray Color = 90
)

var (
	info = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.DefaultText,
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightWhite},
			Text:  " INFO  ",
		},
		Writer: os.Stdout,
	}
	warning = pterm.Warning
	err     = pterm.Error
	debug   = pterm.Debug
	success = pterm.Success
)

func init() {
	pterm.EnableDebugMessages()
	// pterm v0.12.84 no longer bundles gookit/color, so tag rendering is
	// provided by gookit directly. Force color support so color tags render
	// regardless of the detected terminal.
	gookitcolor.ForceColor()
	gookitcolor.Enable = true
	// Temporarily fix output issue including by https://github.com/pterm/pterm/commit/825931aa7ab264074e6c4045c3bdbca5482c758c
	if env.IsTesting() {
		info.Writer = nil
		warning.Writer = nil
		err.Writer = nil
		debug.Writer = nil
		success.Writer = nil
	}
}

// New Functions to create Printer with specific color
func New(color Color) support.Printer {
	return color
}

func Green() support.Printer {
	return New(FgGreen)
}

func Red() support.Printer {
	return New(FgRed)
}

func Blue() support.Printer {
	return New(FgBlue)
}

func Yellow() support.Printer {
	return New(FgYellow)
}

func Cyan() support.Printer {
	return New(FgCyan)
}

func White() support.Printer {
	return New(FgWhite)
}

func Gray() support.Printer {
	return New(FgGray)
}

func Default() support.Printer {
	return New(FgDefault)
}

func Black() support.Printer {
	return New(FgBlack)
}

func Magenta() support.Printer {
	return New(FgMagenta)
}

type Color uint8

func (c Color) Sprint(a ...any) string {
	message := Sprint(a...)
	code := strconv.Itoa(int(c))
	messageLines := strings.Split(message, "\n")
	// Re-apply the color after every inner reset so nested styled segments do
	// not drop the outer color once they close.
	reset := fmt.Sprintf("\x1b[0m\x1b[%sm", code)
	for i, line := range messageLines {
		messageLines[i] = gookitcolor.RenderCode(code, strings.ReplaceAll(line, gookitcolor.ResetSet, reset))
	}
	return strings.Join(messageLines, "\n")
}

func (c Color) Sprintln(a ...any) string {
	return c.Sprint(fmt.Sprintln(a...))
}

func (c Color) Sprintf(format string, a ...any) string {
	return c.Sprint(gookitcolor.Sprintf(format, a...))
}

func (c Color) Sprintfln(format string, a ...any) string {
	return c.Sprint(gookitcolor.Sprintf(format, a...) + "\n")
}

func (c Color) Print(a ...any) *support.Printer {
	Print(c.Sprint(a...))
	p := support.Printer(c)
	return &p
}

func (c Color) Println(a ...any) *support.Printer {
	Print(c.Sprintln(a...))
	p := support.Printer(c)
	return &p
}

func (c Color) Printf(format string, a ...any) *support.Printer {
	Print(c.Sprintf(format, a...))
	p := support.Printer(c)
	return &p
}

func (c Color) Printfln(format string, a ...any) *support.Printer {
	Print(c.Sprintfln(format, a...))
	p := support.Printer(c)
	return &p
}

// Quick use color print message

func Debugf(format string, a ...any) { debug.Print(Sprintf(format, a...)) }

func Debugln(a ...any) { debug.Print(Sprintln(a...)) }

func Errorf(format string, a ...any) { err.Print(Sprintf(format, a...)) }

func Errorln(a ...any) { err.Print(Sprintln(a...)) }

func Infof(format string, a ...any) { info.Print(Sprintf(format, a...)) }

func Infoln(a ...any) { info.Print(Sprintln(a...)) }

func Successf(format string, a ...any) { success.Print(Sprintf(format, a...)) }

func Successln(a ...any) { success.Print(Sprintln(a...)) }

// DEPRECATED: Use Warningf instead
func Warnf(format string, a ...any) { warning.Print(Sprintf(format, a...)) }

// DEPRECATED: Use Warningln instead
func Warnln(a ...any) { warning.Print(Sprintln(a...)) }

func Warningf(format string, a ...any) { warning.Print(Sprintf(format, a...)) }

func Warningln(a ...any) { warning.Print(Sprintln(a...)) }

// StripColor removes color tags and ANSI escape sequences, returning plain text.
func StripColor(s string) string {
	if strings.ContainsRune(s, '<') {
		s = Sprint(s)
	}
	return pterm.RemoveColorFromString(s)
}

// CaptureOutput simulates capturing of os.stdout with a buffer and returns what was written to the screen
func CaptureOutput(f func(w io.Writer)) string {
	var outBuf bytes.Buffer
	info.Writer = &outBuf
	warning.Writer = &outBuf
	err.Writer = &outBuf
	debug.Writer = &outBuf
	success.Writer = &outBuf
	pterm.SetDefaultOutput(&outBuf)
	f(&outBuf)

	content := outBuf.String()
	outBuf.Reset()
	return content
}

// Disable disables color output
func Disable() {
	pterm.DisableColor()
	gookitcolor.Enable = false
}

// Enable enables color output
func Enable() {
	pterm.EnableColor()
	gookitcolor.Enable = true
}
