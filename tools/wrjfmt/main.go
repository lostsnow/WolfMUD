// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// The wrjfmt command can be used to automatically reformat record jar files.
// It's not perfect and has some known issues - see below. Use with caution.
//
// Usage:
//
//	wrjfmt [-t freetext_field] [-i file] [-o file]
//
// The flags are:
//
//	-i file
//		Input file to read (default stdin)
//	-o file
//		Output file to write (default stdout)
//	-t fieldname
//	      The fieldname of the free text section. (default "description")
//
// The output file may be the same as the input file, resulting in the original
// file being overwritten (see warning below as this could cause data loss!).
//
// Known Issues
//
// All comments will be stripped from the output file. In addition all string
// lists will be collapsed. For example:
//
//   OnAction: N
//           : S
//           : E
//           : W
//
// Will be re-written in the output file as:
//
//   OnAction: N : S : E : W
//
// Camel cased field names, such as OnAction, will be title cased: Onaction.
//
// Last of all, named free text fields, e.g. Description, will be turned into
// free text sections.
//
// You have been warned!
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr/ordering"
	"code.wolfmud.org/WolfMUD.git/recordjar"
)

var warning = `
                                  * !WARNING! *

All comments will be stripped from the output file. In addition all string
lists will be collapsed. For example:

  OnAction: N
          : S
          : E
          : W

Will be re-written in the output file as:

  Onaction: N : S : E : W

Camel cased field names, e.g OnAction, will be title cased, e.g. Onaction.

Last of all, named free text fields, e.g. Description, will be turned into free
text sections.

You have been warned!
`

func main() {

	var (
		freetext string   // Name to use for free text section
		in       string   // Name of input file (blank if stdin used)
		r        *os.File // Input file
		out      string   // Name out output file (blank if stdout used)
		w        *os.File // Output file
		err      error
	)

	// Setup flag handling and help text
	fs := flag.NewFlagSet("wrjfmt", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&freetext, "t", "description", "The `fieldname` of the free text section.")
	fs.StringVar(&in, "i", "", "Input `file` to read (default stdin)")
	fs.StringVar(&out, "o", "", "Output `file` to write (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [-t freetext_field] [-i file] [-o file]\n\n", fs.Name())
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nThe output file may be the same as the input file, resulting in the original\nfile being overwritten.\n")
		fmt.Fprintf(fs.Output(), warning)
	}
	fs.Parse(os.Args[1:])

	freetext = strings.ToUpper(freetext)

	switch in {
	case "":
		r = os.Stdin
	default:
		if r, err = os.Open(in); err != nil {
			fmt.Errorf("Error reading input wrj: %w", err)
			return
		}
	}

	jar := recordjar.Read(r, freetext)
	if in != "" {
		r.Close()
	}

	switch out {
	case "":
		w = os.Stdout
	default:
		if w, err = os.Create(out); err != nil {
			fmt.Errorf("Error writing output wrj: %w", err)
			return
		}
		defer w.Close()
	}

	jar.Write(w, freetext, ordering.Attributes)
}
