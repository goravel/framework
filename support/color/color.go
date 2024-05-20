package color

import (
	"github.com/pterm/pterm"
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
	info    = pterm.Info
	warn    = pterm.Warning
	err     = pterm.Error
	debug   = pterm.Debug
	success = pterm.Success
)

type Printer interface {
	Sprint(a ...any) string
	Sprintln(a ...any) string
	Sprintf(format string, a ...any) string
	Sprintfln(format string, a ...any) string
	Print(a ...any) *Printer
	Println(a ...any) *Printer
	Printf(format string, a ...any) *Printer
	Printfln(format string, a ...any) *Printer
}

// New Functions to create Printer with specific color
func New(color Color) Printer {
	return color
}

func Green() Printer {
	return New(FgGreen)
}

func Red() Printer {
	return New(FgRed)
}

func Blue() Printer {
	return New(FgBlue)
}

func Yellow() Printer {
	return New(FgYellow)
}

func Cyan() Printer {
	return New(FgCyan)
}

func White() Printer {
	return New(FgWhite)
}

func Gray() Printer {
	return New(FgGray)
}

func Normal() Printer {
	return New(FgDefault)
}

func Black() Printer {
	return New(FgBlack)
}

func Magenta() Printer {
	return New(FgMagenta)
}

type Color uint8

func (c Color) Sprint(a ...interface{}) string {
	return pterm.Color(c).Sprint(a...)
}

func (c Color) Sprintln(a ...interface{}) string {
	return pterm.Color(c).Sprintln(a...)
}

func (c Color) Sprintf(format string, a ...interface{}) string {
	return pterm.Color(c).Sprintf(format, a...)
}

func (c Color) Sprintfln(format string, a ...interface{}) string {
	return pterm.Color(c).Sprintfln(format, a...)
}

func (c Color) Print(a ...any) *Printer {
	pterm.Color(c).Print(a...)
	p := Printer(c)
	return &p
}

func (c Color) Println(a ...any) *Printer {
	pterm.Color(c).Println(a...)
	p := Printer(c)
	return &p
}

func (c Color) Printf(format string, a ...any) *Printer {
	pterm.Color(c).Printf(format, a...)
	p := Printer(c)
	return &p
}

func (c Color) Printfln(format string, a ...any) *Printer {
	pterm.Color(c).Printfln(format, a...)
	p := Printer(c)
	return &p
}

// Quick use color print message

func Debugln(a ...any) { debug.Println(a...) }

func Errorln(a ...any) { err.Println(a...) }

func Infoln(a ...any) { info.Println(a...) }

func Successln(a ...any) { success.Println(a...) }

func Warnln(a ...any) { warn.Println(a...) }
