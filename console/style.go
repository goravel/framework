package console

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/pterm/pterm"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/convert"
)

var (
	BrandColor = lipgloss.CompleteColor{TrueColor: "#3D8C8D", ANSI256: "30", ANSI: "6"}
	MutedColor = lipgloss.CompleteColor{TrueColor: "#4a4a4a", ANSI256: "240", ANSI: "8"}
	WhiteColor = lipgloss.CompleteColor{TrueColor: "#ffffff", ANSI256: "255", ANSI: "15"}

	DefaultProgressBarStyle   = pterm.NewStyle(pterm.FgLightGreen)
	DefaultProgressTitleStyle = pterm.NewStyle(pterm.FgWhite)

	DefaultTableHeaderColor = BrandColor
	DefaultTableBorderColor = MutedColor

	DefaultTableHeaderStyle  = lipgloss.NewStyle().Foreground(DefaultTableHeaderColor).Bold(true).Padding(0, 1)
	DefaultTableCellStyle    = lipgloss.NewStyle().Padding(0, 1)
	DefaultSpinnerStyle      = lipgloss.NewStyle().Foreground(BrandColor)
	DefaultSpinnerTitleStyle = lipgloss.NewStyle().Foreground(BrandColor)

	DefaultTableStyleFunc = func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return DefaultTableHeaderStyle
		}
		return DefaultTableCellStyle
	}

	DefaultTableOption = console.TableOption{
		Border: lipgloss.RoundedBorder(),

		BorderStyle: lipgloss.NewStyle().Foreground(DefaultTableBorderColor),

		StyleFunc: DefaultTableStyleFunc,

		BorderTop:    convert.Pointer(true),
		BorderBottom: convert.Pointer(true),
		BorderLeft:   convert.Pointer(true),
		BorderRight:  convert.Pointer(true),
		BorderHeader: convert.Pointer(true),
		BorderColumn: convert.Pointer(true),
		BorderRow:    convert.Pointer(false),
	}

	GlobalHuhTheme = func() *huh.Theme {
		t := huh.ThemeCharm()
		t.Focused.Title = t.Focused.Title.Foreground(BrandColor)
		t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(BrandColor)
		t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(BrandColor)
		t.Focused.Base = t.Focused.Base.BorderForeground(BrandColor)

		t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(BrandColor)
		t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(BrandColor)

		t.Focused.FocusedButton = t.Focused.FocusedButton.
			Background(BrandColor).
			Foreground(WhiteColor)

		return t
	}()
)
