package core

import (
	"log"
	"os"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/world/preprocessor"
)

type topic struct {
	Synopsis string
	Usage    []string
	Aliases  []string
	Also     []string
	Examples []string
	Text     string
}

var help map[string]topic

func loadHelp() {
	topics := map[string]topic{}

	log.Printf("Loading help: %s", cfg.helpFile)
	f, err := os.Open(cfg.helpFile)
	if err != nil {
		log.Printf("Error loading help: %s", err)
		help = topics
	}
	defer f.Close()

	jar := recordjar.Read(f, "text")
	preprocessor.Process(jar)

	for _, r := range jar[1:] {
		if _, ok := r["REF"]; ok {
			continue
		}
		ref := decode.Keyword(r["TOPIC"])
		log.Printf("help topic: %s", ref)
		topics[ref] = topic{
			Synopsis: decode.String(r["SYNOPSIS"]),
			Usage:    decode.StringList(r["USAGE"]),
			Aliases:  decode.KeywordList(r["ALIASES"]),
			Also:     decode.KeywordList(r["ALSO"]),
			Examples: decode.StringList(r["EXAMPLES"]),
			Text:     decode.String(r["TEXT"]),
		}
	}

	help = topics
}
