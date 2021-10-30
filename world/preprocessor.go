// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package world

import (
	"bytes"
	"log"
	"regexp"
	"strings"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

// preprocess holds the current state of the preprocessor for a jar.
type preprocess struct {
	recordjar.Jar                    // The jar we are currently processing
	lookup        map[string]int     // A ref to jar index lookup table
	findAtRef     func([]byte) []int // Helper function to find @refs
	recIdx        int                // Index of record in jar being processed
}

// PreProcessor runs the pre-processor on the specified Jar, modifying the
// content of the Jar in the process.
func PreProcessor(j recordjar.Jar) {
	log.Printf("  Pre-processing")

	p := &preprocess{
		Jar:       j,
		findAtRef: regexp.MustCompile("(?:@)(\\w+)(?:\\W|$)").FindSubmatchIndex,
	}
	p.buildRefLookup()
	p.process()
}

// buildRefLookup creates a map for looking up jar record indexes given a
// reference.
func (p *preprocess) buildRefLookup() {
	p.lookup = make(map[string]int, len(p.Jar))
	for x, rec := range p.Jar {
		if key := decode.Keyword(rec["REF"]); key != "" && key[0] != '@' {
			p.lookup[key] = x
		}
	}
}

// process loops through the jar applying preprocessing to each record's fields.
func (p *preprocess) process() {

	var rec recordjar.Record

	for p.recIdx, rec = range p.Jar {
		for field, data := range rec {
			p.Jar[p.recIdx][field] = p.expandAtRef(field, data, "")
		}
	}
}

// expandAtRef expands @ref on a field into the content of the field of the
// same name, but from the record with a ref matching the @ref. An @ref defines
// data at another reference - hence @ref. For example:
//
//  %%
//        Ref: DEFAULT
//      Reset: AFTER→1m JITTER→1m
//    Cleanup: @DEFAULT_EXTRA
//  Inventory: O1 O2
//  %%
//        Ref: DEFAULT_EXTRA
//    Cleanup: AFTER→3m JITTER→4m
//  Inventory: O3 O4
//  %%
//        Ref: M1
//       Name: a bag
//      Alias: BAG
//      Reset: @DEFAULT SPAWN
//    Cleanup: @DEFAULT
//  Inventory: @DEFAULT @DEFAULT_EXTRA
//
//  This is a small bag for carrying things in.
//  %%
//
// The "Reset: @DEFAULT" field for the "Ref: M1" record will cause the
// @DEFAULT to be replaced with the data from the Reset field copied from the
// "Ref: DEFAULT" record.
//
//  - As only the @ref is replaced the field may contain other data - the
//    Reset field on the "Ref: M1" record adds SPAWN after the @DEFAULT.
//
//  - An @ref may reference a field with another @ref - see Cleanup, it
//    references Cleanup at "Ref: DEFAULT" which references Cleanup at
//    "Ref: DEFAULT_EXTRA".
//
//  - A field may contain more than one @ref - see the Inventory field.
//
// The full expansion of M1 after pre-processing would be:
//
//  %%
//        Ref: M1
//       Name: a bag
//      Alias: BAG
//      Reset: AFTER→1m JITTER→1m SPAWN
//    Cleanup: AFTER→3m JITTER→4m
//  Inventory: O1 O2 O3 O4
//
//  This is a small bag for carrying things in.
//  %%
//
func (p *preprocess) expandAtRef(field string, data []byte, seen string) []byte {

	// Quickly exit if no @ref possible
	if len(data) == 0 || bytes.IndexByte(data, '@') == -1 {
		return data
	}

	// Slower quick exit if an @ref not found
	idx := p.findAtRef(data)
	if len(idx) == 0 {
		return data
	}

	ref := decode.Keyword(data[idx[2]:idx[3]])

	p.ifLog(field == "REF", "@ref not allowed on REF fields: @%s", ref)

	// If REF field or @ref already seen (infinite loop), remove @ref from a copy
	// of data and return modified copy.
	if field == "REF" || p.isSeen(field, ref, seen) {
		d := make([]byte, len(data)-(len(ref)+1))
		copy(d, data[:idx[0]])
		copy(d[idx[0]:], data[idx[3]:])
		return d
	}

	var (
		sub        []byte // Substitution text to replace @ref with
		recIdx     int    // Record index in jar @ref references
		recFound   bool   // Record for @ref found in jar?
		fieldFound bool   // Field for @ref found in record?
	)

	// Find @ref within current jar
	if recIdx, recFound = p.lookup[ref]; recFound {
		sub, fieldFound = p.Jar[recIdx][field]
	}

	p.ifLog(!recFound, "@ref record not found, field: %s, @ref: @%s", field, ref)
	p.ifLog(recFound && !fieldFound, "@ref field not found, ref: %s, field: %s", ref, field)

	// expand @ref and store replacement so only expanded once when first seen
	if recFound && fieldFound && !p.isSeen(field, ref, seen) {
		sub = p.expandAtRef(field, sub, seen+"@"+ref+", ")
		p.Jar[recIdx][field] = sub
	}

	// Replace @ref in a copy of the data with its expansion. Expansion may be
	// empty, e.g. ref or field not found, and will cause the @ref to be removed.
	d := make([]byte, len(data)+len(sub)-(len(ref)+1))
	copy(d, data[:idx[0]])
	copy(d[idx[0]:], sub)
	copy(d[idx[0]+len(sub):], data[idx[3]:])

	// Process this field's data again for additional @refs
	return p.expandAtRef(field, d, seen)
}

// ifLog helper to log given message if the test is true, else does nothing.
func (*preprocess) ifLog(test bool, fmt string, arg ...interface{}) {
	if test {
		log.Printf("    "+fmt, arg...)
	}
}

// isSeen helper to scan seen for a ref. Returns true if ref seen else false.
func (p *preprocess) isSeen(field, ref, seen string) bool {
	if seen == "" || strings.Index(seen, "@"+ref+",") == -1 {
		return false
	}

	r := p.Jar[p.recIdx]["REF"] // Top level Jar record we are expanding

	log.Printf("    @ref Loop: %s, field: %s, loop: %s@%s", r, field, seen, ref)
	return true
}
