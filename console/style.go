package console

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/convert"
)

var (
	BrandColor = lipgloss.CompleteColor{TrueColor: "#3D8C8D", ANSI256: "30", ANSI: "6"}
	MutedColor = lipgloss.CompleteColor{TrueColor: "#4a4a4a", ANSI256: "240", ANSI: "8"}

	DefaultTableHeaderColor = BrandColor
	DefaultTableBorderColor = MutedColor

	DefaultTableHeaderStyle = lipgloss.NewStyle().Foreground(DefaultTableHeaderColor).Bold(true).Padding(0, 1)
	DefaultTableCellStyle   = lipgloss.NewStyle().Padding(0, 1)

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
)
