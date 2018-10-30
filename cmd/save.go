// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"log"
	"os"
	"path/filepath"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

// Syntax: SAVE
func init() {
	addHandler(save{}, "SAVE")
}

type save cmd

func (sa save) process(s *state) {

	// Make sure actor is a player
	p := attr.FindPlayer(s.actor)
	if !p.Found() {
		s.msg.Actor.SendBad("You are beyond saving.")
		return
	}

	// Setup account header
	acct := p.(*attr.Player).Account()
	header := acct.Marshal()

	// Save account header and player to Jar
	jar := &recordjar.Jar{}
	*jar = append(*jar, header)
	sa.inventory(jar, s.actor)
	sa.fixInventory(jar)

	// Setup filenames for saving Jar
	acctname := decode.String(header["account"])
	temp := filepath.Join(config.Server.DataDir, "players", acctname+".tmp")
	real := filepath.Join(config.Server.DataDir, "players", acctname+".wrj")

	// Write out player jar to temporary file
	wrj, err := os.Create(temp)
	if err != nil {
		log.Printf("Error saving player: %s, %s", temp, err)
		s.msg.Actor.SendBad("Oops! There was an error saving. Please notify admin.")
		return
	}

	// Set permissions on temporary file
	if config.Server.SetPermissions {
		err = wrj.Chmod(0660)
		if err != nil {
			wrj.Close()
			s.msg.Actor.SendBad("Oops! There was an error saving. Please notify admin.")
			log.Printf("Error changing save file permissions: %s, %s", temp, err)
			return
		}
	}

	// Write out the player file
	jar.Write(wrj, "description")
	wrj.Close()

	// If all went well rename the temporary file to the real file. The rename
	// should be an atomic operation but is dependant on the underlying file
	// system and operating system being used.
	if err := os.Rename(temp, real); err != nil {
		s.msg.Actor.SendBad("Oops! There was an error saving. Please notify admin.")
		log.Printf("Error renaming save file: %s, %s, %s", temp, real, err)
		return
	}

	log.Printf("Player saved: %s.wrj", acctname)
	s.msg.Actor.SendGood("You have been saved.")
	s.ok = true
}

// inventory marshals the passed thing and, if it is an inventory, marshals
// collectable inventory items - recursively.
//
// BUG(diddymus): If a container is not collectable then it and its content
// will not be saved - even if it contains collectable items.
func (sa save) inventory(jar *recordjar.Jar, t has.Thing) {
	*jar = append(*jar, t.(*attr.Thing).Marshal())

	if i := attr.FindInventory(t); i.Found() {
		for _, t := range i.Contents() {
			if t.Collectable() {
				sa.inventory(jar, t)
			}
		}
	}
}

// fixInventory scans the passed jar and fixes "inventory" fields by rewriting
// them to only contain references that are found in the Jar. That is, we drop
// references for non-collectable items that are not saved in the Jar.
func (sa save) fixInventory(jar *recordjar.Jar) {

	// Extract all "ref" fields from the jar
	refs := make(map[string]struct{})
	for _, rec := range *jar {
		refs[string(rec["ref"])] = struct{}{}
	}

	// Find all of the "inventory" fields in the jar and rewrite them to only
	// contain references found in the jar
	for _, rec := range *jar {
		if i, ok := rec["inventory"]; ok {
			newRefs := []string{}
			for _, ref := range decode.KeywordList(i) {
				if _, ok := refs[ref]; ok {
					newRefs = append(newRefs, ref)
				}
			}
			rec["inventory"] = encode.KeywordList(newRefs)
		}
	}
}
