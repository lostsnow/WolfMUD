// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package quota implement per IP address connection rate limiting.
package quota

import (
	"log"
	"math/bits"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
)

// cache records connection attempts and the timestamp of the last connection
// attempt. The attempts field contains bits representing an interval
// cfg.timeslice apart.
var cache = map[string]struct {
	when     time.Time
	attempts uint64
}{}

// quotaPurgeLimit is the maximum quota entries checked in a single purge pass.
// Setting Debug.Quota disables the limit and checks the whole cache.
const quotaPurgeLimit = 1000

// over is a mapping from true/false to '*'/' ' for displaying when over quota.
var over = map[bool]string{true: "*", false: " "}

type pkgConfig struct {
	slots      int
	mask       uint64
	window     time.Duration
	timeslice  time.Duration
	debugQuota bool
	Now        func() time.Time
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config, timeSource func() time.Time) {

	if c.Quota.Slots == 0 || c.Quota.Window == 0 {
		return
	}

	// Max limit of 63 quota slots due to bits in uint64 for mask + 1 extra bit
	slots := c.Quota.Slots
	if slots > 63 {
		log.Printf("WARNING: Limiting Quota.Slots to 63, was %d", slots)
		slots = 63
	}

	cfg = pkgConfig{
		slots:      slots,
		mask:       uint64((1 << (slots + 1)) - 1),
		window:     c.Quota.Window,
		timeslice:  c.Quota.Window / time.Duration(slots),
		debugQuota: c.Debug.Quota,
		Now:        timeSource,
	}

	// Make sure cache is cleared for a new configuration
	for ip := range cache {
		delete(cache, ip)
	}
}

func Status() {
	if cfg.slots == 0 || cfg.window == 0 {
		log.Printf("IP Quota disabled (set Quota.Slots and Quota.Window to enable)")
	} else {
		log.Printf(
			"IP Quota enabled, limit is %d connections per IP address within %s",
			cfg.slots, cfg.window,
		)
	}
}

// Check quota for the given IP address. Check will return true if the IP
// address is over its quota of connections else false.
func Accept(ip string) (allowed bool) {
	if cfg.slots == 0 || cfg.window == 0 {
		return true
	}

	now := cfg.Now()

	c, found := cache[ip]
	if !found {
		c.when, c.attempts = now, 1
		cache[ip] = c
		purge()
		return true
	}

	period := now.Sub(c.when)
	slots := int(period/cfg.timeslice) + 1
	adj := period % cfg.timeslice
	c.when, c.attempts = now.Add(adj), c.attempts<<slots
	c.attempts++
	tries := bits.OnesCount64(c.attempts & cfg.mask)
	cache[ip] = c

	purge()

	return tries <= cfg.slots
}

// purge expired cache entries and delete them. Up to purgeLimit random cache
// entries are checked per call. Number of entries is limited so that time
// spent purging is limited. However, if quota debugging is turned on all cache
// entries are checked. Map access is random and cache will naturally shrink
// over time as entries expire.
func purge() {

	expiry := cfg.Now().Add(-cfg.window)

	limit := 0
	for ip, c := range cache {
		if limit++; limit > quotaPurgeLimit && !cfg.debugQuota {
			break
		}
		if cfg.debugQuota {
			warn := over[bits.OnesCount64(c.attempts&cfg.mask) > cfg.slots]
			log.Printf("QC[%-3d] M[%064b%s] X[%s] A[%s]", limit, c.attempts, warn, c.when.Format("15:04:05"), ip)
		}
		if c.when.Before(expiry) {
			delete(cache, ip)
		}
	}
}

// CacheBits returns the current cache.attempts bits for debugging. It is best
// formatted with '%064b'.
func CacheBits(ip string) uint64 {
	return cache[ip].attempts
}
