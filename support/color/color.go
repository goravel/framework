package color

import (
	"fmt"

	"github.com/pterm/pterm"
)

var (
	green  = pterm.FgGreen
	red    = pterm.FgRed
	blue   = pterm.FgBlue
	yellow = pterm.FgYellow
	cyan   = pterm.FgCyan
	white  = pterm.FgWhite
	gray   = pterm.FgGray
)

var (
	bgGreen  = pterm.BgGreen
	bgRed    = pterm.BgRed
	bgBlue   = pterm.BgBlue
	bgYellow = pterm.BgYellow
	bgCyan   = pterm.BgCyan
	bgWhite  = pterm.BgWhite
	bgGray   = pterm.BgGray
)

var (
	info    = pterm.Info
	warn    = pterm.Warning
	err     = pterm.Error
	debug   = pterm.Debug
	success = pterm.Success
)

func BgBluef(format string, a ...any) { bgBlue.Printf(format, a...) }

func BgBlueln(a ...any) { bgBlue.Println(a...) }

func BgBluep(a ...any) { bgBlue.Print(a...) }

func BgCyanf(format string, a ...any) { bgCyan.Printf(format, a...) }

func BgCyanln(a ...any) { bgCyan.Println(a...) }

func BgCyanp(a ...any) { bgCyan.Print(a...) }

func BgGrayf(format string, a ...any) { bgGray.Printf(format, a...) }

func BgGrayln(a ...any) { bgGray.Println(a...) }

func BgGrayp(a ...any) { bgGray.Print(a...) }

func BgGreenf(format string, a ...any) { bgGreen.Printf(format, a...) }

func BgGreenln(a ...any) { bgGreen.Println(a...) }

func BgGreenp(a ...any) { bgGreen.Print(a...) }

func BgRedf(format string, a ...any) { bgRed.Printf(format, a...) }

func BgRedln(a ...any) { bgRed.Println(a...) }

func BgRedp(a ...any) { bgRed.Print(a...) }

func BgYellowf(format string, a ...any) { bgYellow.Printf(format, a...) }

func BgYellowln(a ...any) { bgYellow.Println(a...) }

func BgYellowp(a ...any) { bgYellow.Print(a...) }

func BgWhitef(format string, a ...any) { bgWhite.Printf(format, a...) }

func BgWhiteln(a ...any) { bgWhite.Println(a...) }

func BgWhitep(a ...any) { bgWhite.Print(a...) }

func Bluef(format string, a ...any) { blue.Printf(format, a...) }

func Blueln(a ...any) { blue.Println(a...) }

func Bluep(a ...any) { blue.Print(a...) }

func Cyanf(format string, a ...any) { cyan.Printf(format, a...) }

func Cyanln(a ...any) { cyan.Println(a...) }

func Cyanp(a ...any) { cyan.Print(a...) }

func Grayf(format string, a ...any) { gray.Printf(format, a...) }

func Grayln(a ...any) { gray.Println(a...) }

func Grayp(a ...any) { gray.Print(a...) }

func Greenf(format string, a ...any) { green.Printf(format, a...) }

func Greenln(a ...any) { green.Println(a...) }

func Greenp(a ...any) { green.Print(a...) }

func Print(a ...any) { fmt.Print(a...) }

func Printf(format string, a ...any) { fmt.Printf(format, a...) }

func Println(a ...any) { fmt.Println(a...) }

func Redf(format string, a ...any) { red.Printf(format, a...) }

func Redln(a ...any) { red.Println(a...) }

func Redp(a ...any) { red.Print(a...) }

func Yellowf(format string, a ...any) { yellow.Printf(format, a...) }

func Yellowln(a ...any) { yellow.Println(a...) }

func Yellowp(a ...any) { yellow.Print(a...) }

func Whitef(format string, a ...any) { white.Printf(format, a...) }

func Whiteln(a ...any) { white.Println(a...) }

func Whitep(a ...any) { white.Print(a...) }

func Sbluef(format string, a ...any) string { return blue.Sprintf(format, a...) }

func Sblueln(a ...any) string { return blue.Sprintln(a...) }

func Sbluep(a ...any) string { return blue.Sprint(a...) }

func Scyanf(format string, a ...any) string { return cyan.Sprintf(format, a...) }

func Scyanln(a ...any) string { return cyan.Sprintln(a...) }

func Scyanp(a ...any) string { return cyan.Sprint(a...) }

func Sgrayf(format string, a ...any) string { return gray.Sprintf(format, a...) }

func Sgrayln(a ...any) string { return gray.Sprintln(a...) }

func Sgrayp(a ...any) string { return gray.Sprint(a...) }

func Sgreenf(format string, a ...any) string { return green.Sprintf(format, a...) }

func Sgreenln(a ...any) string { return green.Sprintln(a...) }

func Sgreenp(a ...any) string { return green.Sprint(a...) }

func Sprintf(format string, a ...any) string { return fmt.Sprintf(format, a...) }

func Sprintln(a ...any) string { return fmt.Sprintln(a...) }

func Sprint(a ...any) string { return fmt.Sprint(a...) }

func Sredf(format string, a ...any) string { return red.Sprintf(format, a...) }

func Sredln(a ...any) string { return red.Sprintln(a...) }

func Sredp(a ...any) string { return red.Sprint(a...) }

func Syellowf(format string, a ...any) string { return yellow.Sprintf(format, a...) }

func Syellowln(a ...any) string { return yellow.Sprintln(a...) }

func Syellowp(a ...any) string { return yellow.Sprint(a...) }

func Swhitef(format string, a ...any) string { return white.Sprintf(format, a...) }

func Swhiteln(a ...any) string { return white.Sprintln(a...) }

func Swhitep(a ...any) string { return white.Sprint(a...) }

// Quick use color print message

func Debugf(format string, a ...any) { debug.Printf(format, a...) }

func Debugln(a ...any) { debug.Println(a...) }

func Debugp(a ...any) { debug.Print(a...) }

func Errorf(format string, a ...any) { err.Printf(format, a...) }

func Errorln(a ...any) { err.Println(a...) }

func Errorp(a ...any) { err.Print(a...) }

func Infof(format string, a ...any) { info.Printf(format, a...) }

func Infoln(a ...any) { info.Println(a...) }

func Infop(a ...any) { info.Print(a...) }

func Successf(format string, a ...any) { success.Printf(format, a...) }

func Successln(a ...any) { success.Println(a...) }

func Successp(a ...any) { success.Print(a...) }

func Warnf(format string, a ...any) { warn.Printf(format, a...) }

func Warnln(a ...any) { warn.Println(a...) }

func Warnp(a ...any) { warn.Print(a...) }
