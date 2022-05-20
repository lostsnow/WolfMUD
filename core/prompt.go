package core

import (
	"fmt"
	"sort"
	"strings"

	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Constants for the predefined prompt styles.
const (
	PromptStyleUnset   = ""
	PromptStyleNone    = "NONE"    // \n
	PromptStyleCursor  = "CURSOR"  // \n>
	PromptStyleMinimal = "MINIMAL" // \nHc␠
	PromptStyleBrief   = "BRIEF"   // \nH:c>
	PromptStyleShort   = "SHORT"   // \nH:c/m>
	PromptStyleLong    = "LONG"    // \nHealth: c/m>
)

// Prompt is a map of functions for building each of the prompt styles. Once
// created the map should not be modified. Example usage:
//
//  Prompt[s.actor.As[PromptStyle]](s.actor)
//
var Prompt = map[string]func(who *Thing){

	PromptStyleUnset: func(*Thing) {
		// Handles lookup when As[Prompt] not set
	},

	PromptStyleNone: func(who *Thing) {
		mailbox.Suffix(who.As[UID], "\n"+text.Magenta)
	},

	PromptStyleCursor: func(who *Thing) {
		mailbox.Suffix(who.As[UID], "\n"+text.Magenta+">")
	},

	PromptStyleMinimal: func(who *Thing) {
		mailbox.Suffix(who.As[UID], fmt.Sprintf(
			"\n"+text.Blue+"H%d"+text.Magenta+"␠", who.Int[HealthCurrent],
		))
	},

	PromptStyleBrief: func(who *Thing) {
		mailbox.Suffix(who.As[UID], fmt.Sprintf(
			"\n"+text.Blue+"H:%d"+text.Magenta+">", who.Int[HealthCurrent],
		))
	},

	PromptStyleShort: func(who *Thing) {
		mailbox.Suffix(who.As[UID], fmt.Sprintf(
			"\n"+text.Blue+"H:%[1]d/%[2]d"+text.Magenta+">",
			who.Int[HealthCurrent], who.Int[HealthMaximum],
		))
	},

	PromptStyleLong: func(who *Thing) {
		mailbox.Suffix(who.As[UID], fmt.Sprintf(
			"\n"+text.Blue+"Health: %[1]d/%[2]d"+text.Magenta+">",
			who.Int[HealthCurrent], who.Int[HealthMaximum],
		))
	},
}

// PromptList returns a comma+space separated list of available prompt styles
// as a string.
var PromptList = func() string {
	list := []string{}
	for x := range Prompt {
		if x == PromptStyleUnset {
			continue
		}
		list = append(list, x)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}()
