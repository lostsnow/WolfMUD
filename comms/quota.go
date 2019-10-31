// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"log"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
)

// TimeSource is a function that returns the current time as a time.Time. This
// is typically time.Now but may be replaced with a fake time for testing.
type TimeSource func() time.Time

// quota is used to implement per IP connection quotas.
type quota struct {
	Now      TimeSource      // Time source for current time
	cache    map[string]Ring // tracking map keyed by IP
	window   int64           // cache expiry period if not over quota
	timeout  int64           // cache expiry if over quota
	stats    int64           // Minimum reporting period for quota stats
	sweepDue int64           // Time after which next idle sweep is due
	statsDue int64           // Time after which stats are displayed
}

// NewQuota returns a new, initialised connection quota. Connection quota can
// either be per listening server port or per server. The TimeSource is a
// function returning the current time as a time.Time, typically time.Now.
func NewQuota(ts TimeSource) *quota {
	return &quota{
		Now:      ts,
		cache:    make(map[string]Ring),
		window:   config.Quota.Window.Nanoseconds(),
		timeout:  config.Quota.Timeout.Nanoseconds(),
		stats:    config.Quota.Stats.Nanoseconds(),
		sweepDue: config.Quota.Window.Nanoseconds() + ts().UnixNano(),
		statsDue: config.Quota.Stats.Nanoseconds() + ts().UnixNano(),
	}
}

// Enabled returns true if connection quotas are enabled, else false. Quotas
// can be disabled by setting Quota.Window in the configuration file to 0.
func (q *quota) Enabled() bool {
	return q.window != 0
}

// Quota records a connection attempt by an IP address and returns true if the
// IP address is currently over its quota, based on prior connection rates to
// the server, otherwise false.
//
//
// Operating Mode One
//
// Mode one is selected when the configuration value of Quota.Timeout specifies
// a duration that is not zero.
//
// An IP address has a quota of connection attempts it can use within the
// Quota.Window period. If an IP address exceeds its allocated quota within the
// Quota.Window period then Quota will return true (over quota) until the
// Quota.Timeout period has passed.
//
// Example configuration file entries:
//
//  Quota.Window:  10s
//  Quota.Timeout: 30m
//
// In this example if an IP address exceeds its quota of connection attempts
// within 10 seconds then Quota will return true for the IP address for the
// next 30 minutes. The 30 minutes is regardless of whether the client still
// tries to connect to the server or not.
//
//
// Operating Mode Two
//
// Mode two is selected when the configuration value of Quota.Timeout specifies
// a zero duration.
//
// If the Quota.Timeout period is set to 0 then Quota will return true (over
// quota) until a waiting period equal to Quota.Window has passed, with no
// further connection attempts being made by the IP address. If the over quota
// IP address tries to connect to the server during the waiting period then the
// waiting period will be reset to Quota.Window again.
//
// Example configuration file entries:
//
//  Quota.Window: 30s
//  Quota.Timeout:  0
//
// In this example if an IP address exceeds its quota of connection attempts
// within 30 seconds then quota will return true for the IP address until the
// IP address makes no further connection attempts for 30 seconds. If the IP
// address tries to connect to the server while banned the waiting period of 30
// seconds will start over again from the time of the connection attempt.
//
//
// IP Addresses AND Connection Quotas
//
// The connection quota for an IP address is set via the compile time constant
// ringSize, set in ring.go - currently set to 4.
//
// This does not mean that each IP address is limited to only 4 connections to
// the server. It limits the number of initial connections in a short space of
// time. This is to protect the server against DOS style attacks caused by
// clients connecting and dropping connections. For example by using:
//
//  seq 1000000 | xargs -n1 -i@ netcat -z 127.0.0.1 4001
//
// A better solution would be to use a proper firewall and set up connection
// and rate limiting. However some users want some basic functionality built
// into the server.
func (q *quota) Quota(ip string) (overQuota bool) {

	if q.window == 0 {
		return
	}

	now := q.Now().UnixNano()
	c := q.cache[ip]

	// Clear any expired quota for current IP address
	for v := c.Last(); !c.Empty() && now > v; v = c.Last() {
		c.Popd()
	}

	// If cache for IP address is full the IP is over its quota
	if c.Full() {

		overQuota = true

		// Set first/last cache entries to delay expiry and delay idle sweep
		if q.timeout != 0 {
			if c.First() != c.Last() {
				c.FirstReplace(now + q.timeout)
				c.LastReplace(now + q.timeout)
			}
		} else {
			c.FirstReplace(now + q.window)
			c.LastReplace(now + q.window)
		}
	} else {
		c.Unshift(now + q.window)
	}

	q.cache[ip] = c

	// Run an idle sweep if current IP not over its quota and a sweep is due. If
	// IP address is over its quota we want to handle it as quickly as possible
	// so we ignore any idle sweep in that case.
	if !overQuota && now > q.sweepDue {
		q.CacheSweep()
	}

	return
}

// CacheSweep iterates over the quota cache and evicts idle cache entries that
// have expired. CacheSweep will log statistics after the Quote.Stats period
// has passed. Setting Quote.Stats to 0 will disable statistics reporting.
func (q *quota) CacheSweep() {
	entries, stale, banned := len(q.cache), 0, 0
	now := q.Now().UnixNano()
	for x, c := range q.cache {
		f := c.First()
		if now > f {
			delete(q.cache, x)
			stale++
		} else {
			if c.Full() && f == c.Last() {
				banned++
			}
		}
	}

	if q.stats != 0 && now > q.statsDue {
		log.Printf(
			"quota: %d entries [%d evicted, %d over quota, %d recent]",
			entries, stale, banned, entries-stale-banned,
		)
		q.statsDue = now + q.stats
	}

	// Schedule next idle sweep to be twice frequency of the window period. This
	// avoids over scheduling but keeps the cache small.
	q.sweepDue = now + q.window + q.window
}

// Query allows the cached state of an IP address to be checked. Query will
// return the current number of quota used, whether the IP address is over
// quota and whether the IP address quota are expired. If there is no cache for
// the specified IP address then count will be -1, overQuota false and expired
// false.
func (q *quota) Query(ip string) (count int, overQuota, expired bool) {
	c, ok := q.cache[ip]
	if !ok {
		return -1, false, false
	}
	count = c.Len()
	overQuota = c.Full() && (c.First() == c.Last())
	expired = (count > 0) && (q.Now().UnixNano() > c.First())
	return
}
