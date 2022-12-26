package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/text"
	"code.wolfmud.org/WolfMUD.git/world/preprocessor"
)

// help contains a formatted and cross-indexed library.
var help *library

// library contains help pages and a cross-reference. Pages are indexed by the
// help topic. The corss-reference maps all aliases to the help topic
// reference. To check if a page exists and to get the main topic reference:
// help.xref[ref], and to retrieve the help page: help.pages[help.xref[ref]].
type library struct {
	pages map[string]string // help pages by topic reference
	xref  map[string]string // aliases to topic cross-reference
}

// build is used to hold all of the help topics and indexing while building the
// library.
type build struct {
	topics map[string]topic
	xref   map[string]string
	groups map[string][]string
	pages  map[string]string
	xpad   string
}

// topic represents a single help topic in the build.
type topic struct {
	Groups   []string
	Synopsis string
	Usage    []string
	Aliases  []string
	Also     []string
	Examples []string
	Text     string
}

// loadLibrary loads the help *.wrj files into the library.
func loadLibrary(path string) {
	topics := map[string]topic{}
	xref := map[string]string{}
	groups := map[string][]string{}
	pages := map[string]string{}
	maxLen := 0

	log.Printf("Loading help: %s", path)
	f, err := os.Open(cfg.helpFile)
	if err != nil {
		log.Printf("Error loading help: %s", err)
	}
	defer f.Close()

	jar := recordjar.Read(f, "text")
	preprocessor.Process(jar)

	if len(jar) == 0 || decode.Boolean(jar[0]["DISABLED"]) == true {
		log.Print("Help system has been disabled.")
		return
	}

	for _, r := range jar[1:] {
		if _, ok := r["REF"]; ok {
			continue
		}
		ref := decode.Keyword(r["TOPIC"])
		if len(ref) > maxLen {
			maxLen = len(ref)
		}
		topics[ref] = topic{
			Groups:   decode.KeywordList(r["GROUP"]),
			Synopsis: decode.String(r["SYNOPSIS"]),
			Usage:    decode.StringList(r["USAGE"]),
			Aliases:  decode.KeywordList(r["ALIASES"]),
			Also:     decode.KeywordList(r["ALSO"]),
			Examples: decode.StringList(r["EXAMPLES"]),
			Text:     decode.String(r["TEXT"]),
		}
		xref[ref] = ref
		for _, a := range topics[ref].Aliases {
			xref[a] = ref
		}
		for _, a := range topics[ref].Groups {
			groups[a] = append(groups[a], ref)
		}
	}

	b := &build{topics, xref, groups, pages, strings.Repeat("␠", maxLen+2)}
	log.Print("  Formatting help pages")
	b.format()
	help = &library{pages, xref}
	b = nil
	log.Printf("Loaded and formatted %d help pages.", len(help.pages))
}

// format topics as pages for display.
func (b build) format() {
	page := &strings.Builder{}

	for ref, t := range b.topics {

		fmt.Fprintf(page, "%sTopic: %s%s", text.Cyan, text.Reset, ref)
		if len(t.Aliases) > 0 {
			fmt.Fprintf(page, " (aliases: %s%s)", text.Reset, text.List(t.Aliases))
		}
		if len(t.Synopsis) > 0 {
			fmt.Fprintf(page, "%s\n       %s", text.Reset, t.Synopsis)
		}
		if len(t.Usage) > 0 {
			fmt.Fprintf(page, "%s\n\nUsage: %s%s",
				text.Cyan, text.Reset, strings.Join(t.Usage, "\n       "))
		}
		if len(t.Text) > 0 {
			fmt.Fprintf(page, "%s\n\n%s", text.Reset, t.Text)
		}
		if len(b.groups[ref]) > 0 {
			fmt.Fprintf(page, "\n%s", text.Reset)
			for _, ref := range b.groups[ref] {
				fmt.Fprintf(page, "\n%s%s␠␠%s",
					b.xpad[len(ref)*len("␠"):], ref, b.topics[ref].Synopsis,
				)
			}
		}
		if len(t.Examples) > 0 {
			fmt.Fprintf(page, "%s\n\nExamples: %s%s",
				text.Cyan, text.Reset, strings.Join(t.Examples, "\n          "),
			)
		}
		if len(t.Also) > 0 {
			fmt.Fprintf(page, "%s\n\nSee also: %s%s",
				text.Cyan, text.Reset, text.List(t.Also),
			)
		}

		b.pages[ref] = page.String()
		page.Reset()
	}
}
