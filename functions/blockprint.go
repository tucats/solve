package functions

import (
	"sort"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
	"github.com/tucats/ego/util"
)

var fontSet []string

func blockPrint(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	initFonts()

	msg := util.GetString(args[0])

	fontName := "standard"
	if len(args) > 1 {
		fontName = util.GetString(args[1])
	}

	if !isFont(fontName) {
		return nil, errors.New(errors.ErrNoSuchAsset).Context(fontName)
	}

	myFigure := figure.NewFigure(msg, fontName, true)

	return myFigure.String(), nil
}

func blockFonts(s *symbols.SymbolTable, args []interface{}) (interface{}, *errors.EgoError) {
	initFonts()

	result := datatypes.NewArray(datatypes.StringType, len(fontSet))

	for idx, name := range fontSet {
		_ = result.Set(idx, name)
	}

	return result, nil
}

func initFonts() {
	if fontSet == nil {
		fontSet = figure.AssetNames()
		sort.Strings(fontSet)

		for idx := 0; idx < len(fontSet); idx++ {
			fontSet[idx] = strings.TrimPrefix(strings.TrimSuffix(fontSet[idx], ".flf"), "fonts/")
		}
	}
}

func isFont(name string) bool {
	for _, font := range fontSet {
		if strings.EqualFold(name, font) {
			return true
		}
	}

	return false
}
