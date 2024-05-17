package color

import (
	"fmt"

	"github.com/pterm/pterm"
)

var (
	Green  = pterm.FgGreen
	Red    = pterm.FgRed
	Blue   = pterm.FgBlue
	Yellow = pterm.FgYellow
	Cyan   = pterm.FgCyan
	White  = pterm.FgWhite
	Gray   = pterm.FgGray
)

var (
	BgGreen  = pterm.BgGreen
	BgRed    = pterm.BgRed
	BgBlue   = pterm.BgBlue
	BgYellow = pterm.BgYellow
	BgCyan   = pterm.BgCyan
	BgWhite  = pterm.BgWhite
	BgGray   = pterm.BgGray
)

var (
	Info    = pterm.Info
	Warn    = pterm.Warning
	Error   = pterm.Error
	Debug   = pterm.Debug
	Success = pterm.Success
)

func BgBluef(format string, a ...any) { BgBlue.Printf(format, a...) }

func BgBlueln(a ...any) { BgBlue.Println(a...) }

func BgBluep(a ...any) { BgBlue.Print(a...) }

func BgCyanf(format string, a ...any) { BgCyan.Printf(format, a...) }

func BgCyanln(a ...any) { BgCyan.Println(a...) }

func BgCyanp(a ...any) { BgCyan.Print(a...) }

func BgGrayf(format string, a ...any) { BgGray.Printf(format, a...) }

func BgGrayln(a ...any) { BgGray.Println(a...) }

func BgGrayp(a ...any) { BgGray.Print(a...) }

func BgGreenf(format string, a ...any) { BgGreen.Printf(format, a...) }

func BgGreenln(a ...any) { BgGreen.Println(a...) }

func BgGreenp(a ...any) { BgGreen.Print(a...) }

func BgRedf(format string, a ...any) { BgRed.Printf(format, a...) }

func BgRedln(a ...any) { BgRed.Println(a...) }

func BgRedp(a ...any) { BgRed.Print(a...) }

func BgYellowf(format string, a ...any) { BgYellow.Printf(format, a...) }

func BgYellowln(a ...any) { BgYellow.Println(a...) }

func BgYellowp(a ...any) { BgYellow.Print(a...) }

func BgWhitef(format string, a ...any) { BgWhite.Printf(format, a...) }

func BgWhiteln(a ...any) { BgWhite.Println(a...) }

func BgWhitep(a ...any) { BgWhite.Print(a...) }

func Bluef(format string, a ...any) { Blue.Printf(format, a...) }

func Blueln(a ...any) { Blue.Println(a...) }

func Bluep(a ...any) { Blue.Print(a...) }

func Cyanf(format string, a ...any) { Cyan.Printf(format, a...) }

func Cyanln(a ...any) { Cyan.Println(a...) }

func Cyanp(a ...any) { Cyan.Print(a...) }

func Grayf(format string, a ...any) { Gray.Printf(format, a...) }

func Grayln(a ...any) { Gray.Println(a...) }

func Grayp(a ...any) { Gray.Print(a...) }

func Greenf(format string, a ...any) { Green.Printf(format, a...) }

func Greenln(a ...any) { Green.Println(a...) }

func Greenp(a ...any) { Green.Print(a...) }

func Print(a ...any) { fmt.Print(a...) }

func Printf(format string, a ...any) { fmt.Printf(format, a...) }

func Println(a ...any) { fmt.Println(a...) }

func Redf(format string, a ...any) { Red.Printf(format, a...) }

func Redln(a ...any) { Red.Println(a...) }

func Redp(a ...any) { Red.Print(a...) }

func Yellowf(format string, a ...any) { Yellow.Printf(format, a...) }

func Yellowln(a ...any) { Yellow.Println(a...) }

func Yellowp(a ...any) { Yellow.Print(a...) }

func Whitef(format string, a ...any) { White.Printf(format, a...) }

func Whiteln(a ...any) { White.Println(a...) }

func Whitep(a ...any) { White.Print(a...) }

func Sbluef(format string, a ...any) string { return Blue.Sprintf(format, a...) }

func Sblueln(a ...any) string { return Blue.Sprintln(a...) }

func Sbluep(a ...any) string { return Blue.Sprint(a...) }

func Scyanf(format string, a ...any) string { return Cyan.Sprintf(format, a...) }

func Scyanln(a ...any) string { return Cyan.Sprintln(a...) }

func Scyanp(a ...any) string { return Cyan.Sprint(a...) }

func Sgrayf(format string, a ...any) string { return Gray.Sprintf(format, a...) }

func Sgrayln(a ...any) string { return Gray.Sprintln(a...) }

func Sgrayp(a ...any) string { return Gray.Sprint(a...) }

func Sgreenf(format string, a ...any) string { return Green.Sprintf(format, a...) }

func Sgreenln(a ...any) string { return Green.Sprintln(a...) }

func Sgreenp(a ...any) string { return Green.Sprint(a...) }

func Sprintf(format string, a ...any) string { return fmt.Sprintf(format, a...) }

func Sprintln(a ...any) string { return fmt.Sprintln(a...) }

func Sprint(a ...any) string { return fmt.Sprint(a...) }

func Sredf(format string, a ...any) string { return Red.Sprintf(format, a...) }

func Sredln(a ...any) string { return Red.Sprintln(a...) }

func Sredp(a ...any) string { return Red.Sprint(a...) }

func Syellowf(format string, a ...any) string { return Yellow.Sprintf(format, a...) }

func Syellowln(a ...any) string { return Yellow.Sprintln(a...) }

func Syellowp(a ...any) string { return Yellow.Sprint(a...) }

func Swhitef(format string, a ...any) string { return White.Sprintf(format, a...) }

func Swhiteln(a ...any) string { return White.Sprintln(a...) }

func Swhitep(a ...any) string { return White.Sprint(a...) }

// Quick use color print message

func Debugf(format string, a ...any) { Debug.Printf(format, a...) }

func Debugln(a ...any) { Debug.Println(a...) }

func Debugp(a ...any) { Debug.Print(a...) }

func Errorf(format string, a ...any) { Error.Printf(format, a...) }

func Errorln(a ...any) { Error.Println(a...) }

func Errorp(a ...any) { Error.Print(a...) }

func Infof(format string, a ...any) { Info.Printf(format, a...) }

func Infoln(a ...any) { Info.Println(a...) }

func Infop(a ...any) { Info.Print(a...) }

func Successf(format string, a ...any) { Success.Printf(format, a...) }

func Successln(a ...any) { Success.Println(a...) }

func Successp(a ...any) { Success.Print(a...) }

func Warnf(format string, a ...any) { Warn.Printf(format, a...) }

func Warnln(a ...any) { Warn.Println(a...) }

func Warnp(a ...any) { Warn.Print(a...) }
