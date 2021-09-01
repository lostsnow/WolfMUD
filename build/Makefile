# Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
#
# Use of this source code is governed by the license in the LICENSE file
# included with the source code.
#
# Makefile to build a WolfMUD release. Targets of note:
#
#   all - make all platforms (default)
#
# Builds should be reproduce able for any given Go version and build date.
# Build date can be specified using the 'now' environment variable.
#
SHELL := /bin/bash
info := $(shell ls ../[A-Z][A-Z]*)
tars := linux-386 linux-amd64 linux-arm5 linux-arm6 linux-arm7
zips := windows-386 windows-amd64
version := $(shell git describe --dirty --always)
goVersion := $(shell go version)

# Don't build with a release with a dirty worktree otherwise the source
# archives will not match the sources used to build the binaries.
ifeq ($(patsubst %dirty,,$(lastword $(version))),)
$(error Git working tree is dirty, Cannot build release)
endif

# For reproduceable builds we can pass in a specific date, also relies on
# version of Go being used to compile.
ifndef now
	now := $(shell date +@%s)
endif
nowDate := $(shell ./niceDate "$(now)")
tarDate := $(shell date -u -d"$(now)" +%Y-%m-%d\ 00:00Z)
zipDate := $(shell date -u -d"$(now)" +%Y%m%d0000.00)

.PHONY: all
all: | $(tars) $(zips) source index.txt finish jt

$(tars): %: WolfMUD-%.tgz WolfMUD-%.tgz.txt
$(zips): %: WolfMUD-%.zip WolfMUD-%.zip.txt

# Note 'jt' is the "Just Text" static site generator used for the WolfMUD.org
# website and is not (yet!) generally available. It is used to turn the text
# files into a linked webpage for the downloads section of the website. In its
# absence an unlinked version named 'full-index.txt' is created instead.
.PHONY: jt
jt:
	cd $(version) ;\
	which "jt" && jt ./index.txt || \
	cat index.txt WolfMUD-*.txt \
	| sed -e '/^\(link:.*\)$$/{h;d}' -e '/^Size..:/{G}' \
	| sed -e 's/^link:\(.*\)/Link..:\1/' > full-index.txt

index.txt:
	echo "Title: $(version)" > $@ ;\
	echo "Abstract: Released: $(nowDate)" >> $@ ;\
	echo "Tags: $(version)" >> $@ ;\
	echo "jtconv: nofold" >> $@ ;\
	echo "" >> $@ ;\
	echo "WolfMUD: $(version)" | ./center >> $@ ;\
	echo "" >> $@ ;\
	echo "Released: $(nowDate)" | ./center >> $@ ;\
	echo "Compiled with: $(goVersion)" | ./center >> $@

WolfMUD-%.txt:
	echo "link: WolfMUD-$*" > $@ ;\
	echo "" >> $@ ;\
	pcregrep -o1 '(?:$*:) (.*)' descriptions.txt >> $@ ;\
	echo "MD5...: `md5sum WolfMUD-$* | cut -d\  -f1`" >> $@ ;\
	echo "SHA256: `sha256sum WolfMUD-$* | cut -d\  -f1`" >> $@ ;\
	echo "Size..: ~`stat -c %s WolfMUD-$* | numfmt --to=iec`" >> $@

WolfMUD-%.tgz: bin/% $(info) docs config zones players
	tar --mtime="${tarDate}" --sort=name --owner=0 --group=0 --numeric-owner -zcf $@ WolfMUD ;\
	find WolfMUD -type f -executable -delete

WolfMUD-%.zip: bin/% $(info) docs config zones players
	find WolfMUD -exec touch -t $(zipDate) -m {} \; ;\
	zip -Xq9r $@ WolfMUD ;\
	find WolfMUD -type f -executable -delete

.PHONY: bin/%
bin/%: WolfMUD/bin
	$(MAKE) --directory=.. $(@F) ;\
	cp ../$@/* ./WolfMUD/bin/

WolfMUD:
	mkdir -p WolfMUD
WolfMUD/bin: WolfMUD
	mkdir -p WolfMUD/bin
WolfMUD/docs: WolfMUD
	mkdir -p WolfMUD/docs
WolfMUD/data: WolfMUD
	mkdir -p WolfMUD/data
WolfMUD/data/zones: WolfMUD/data
	mkdir -p WolfMUD/data/zones
WolfMUD/data/players: WolfMUD/data
	mkdir -p WolfMUD/data/players

$(info): WolfMUD
	cp $@ ./WolfMUD/

docs: WolfMUD/docs
	cp ../docs/*.txt ./WolfMUD/docs/

config: WolfMUD/data
	cp ../data/config.wrj ./WolfMUD/data/

zones: WolfMUD/data/zones
	cp ../data/zones/*.wrj ./WolfMUD/data/zones/

players: WolfMUD/data/players

source: | WolfMUD-source.tgz WolfMUD-source.tgz.txt
source: | WolfMUD-source.zip WolfMUD-source.zip.txt

WolfMUD-source.tgz:
	cd .. ;\
	git archive --format=tgz -o ./build/WolfMUD-source.tgz --prefix=WolfMUD/ $(version)

WolfMUD-source.zip:
	cd .. ;\
	git archive --format=zip -o ./build/WolfMUD-source.zip --prefix=WolfMUD/ $(version)

.PHONY: finish
finish:
	[ -d "./$(version)" ] && rm -rf ./$(version) ;\
	mkdir $(version) ;\
	mv WolfMUD-* $(version)/ ;\
	mv index.txt $(version)/ ;\
	rm -rf WolfMUD/
