package lib

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func sortModsByName(mods []ModEntry) {
	sort.Slice(mods, func(i, j int) bool {
		a := mods[i].Name
		b := mods[j].Name

		aFirst := []rune(a)[0]
		bFirst := []rune(b)[0]

		aIsSpecial := !unicode.IsLetter(aFirst) && !unicode.IsNumber(aFirst)
		bIsSpecial := !unicode.IsLetter(bFirst) && !unicode.IsNumber(bFirst)

		if aIsSpecial != bIsSpecial {
			return aIsSpecial
		}
		return strings.ToLower(a) < strings.ToLower(b)
	})
}

func PrintMods(db *DB) {
	sortModsByName(db.Mods)

	// ищем самое длинное CodeName
	maxCodeLen := 0
	for _, m := range db.Mods {
		if len(m.CodeName) > maxCodeLen {
			maxCodeLen = len(m.CodeName)
		}
	}
	codeColWidth := maxCodeLen + 4

	fmt.Printf("%-9s %-*s %s\n", "Enabled", codeColWidth, "CodeName", "Name")

	for _, m := range db.Mods {
		enabled := "❌"
		if m.Enabled {
			enabled = "✅"
		}
		fmt.Printf("%-8s %-*s %s\n", enabled, codeColWidth, m.CodeName, m.Name)
	}
}
