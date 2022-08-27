package debugger

import (
	"fmt"

	"github.com/tucats/ego/app-cli/tables"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/errors"
)

const defaultHelpIndent = 3

var helpText = [][]string{
	{"break at <line>", "Halt execution at a given line number"},
	{"break when <expression>", "Halt execution when expression is true"},
	{"break clear at <line>", "Remove breakpoint for line number"},
	{"break clear when <expression>", "Remove breakpoint for expression"},
	{"break load [\"file\"]", "Load breakpoints from named file"},
	{"break save [\"file\"]", "Save breakpoint list to named file"},
	{"continue", "Resume execution of the program"},
	{"exit", "Exit the debugger"},
	{"help", "display this help text"},
	{"print", "Print the value of an expression"},
	{"set <variable> = <expression>", "Set a variable to a value"},
	{"show breaks", "Display list of breakpoints"},
	{"show calls [<count>]", "Display the call stack to the given depth"},
	{"show symbols", "Display the current symbol table"},
	{"show line", "Display the current program line"},
	{"show scope", "Display nested call scope"},
	{"show source [start [:end]]", "Display source of current module"},
	{"step [into]", "Execute the next line of the program"},
	{"step over", "Step over a function call to the next line in this program"},
	{"step return", "Execute until the next return operation"},
}

func Help() *errors.EgoError {
	table, err := tables.New([]string{"Command", "Description"})

	for _, helpItem := range helpText {
		err = table.AddRow(helpItem)
	}

	if errors.Nil(err) {
		fmt.Println("Debugger commands:")

		_ = table.ShowUnderlines(false).ShowHeadings(false).SetIndent(defaultHelpIndent)
		_ = table.SetOrderBy("Command")
		_ = table.Print(ui.TextFormat)
	}

	return err
}
