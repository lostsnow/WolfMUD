// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms_test

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.wolfmud.org/WolfMUD.git/comms"
	"code.wolfmud.org/WolfMUD.git/config"
)

// Constants for use with veryify to make tests more readable
const (
	UnderQuota  = false
	OverQuota   = true
	Valid       = false
	Expired     = true
	IPNotCached = -1
)

// maxQuota is the maximum number of Quota an IP address can have within a
// config.Quota.Window period before going over quota.
var maxQuota = (&comms.Ring{}).Cap()

// verify is a helper that checks the status of the quota cache for an IP
// address. As quota is not exported we pass in the Query method as a method
// value. Messy, but it works.
func verify(t *testing.T,
	Query func(string) (int, bool, bool),
	ip string, count int, over, expired bool,
) {
	t.Helper()
	c, o, e := Query(ip)
	if c != count {
		t.Errorf("%q count, have: %d, want: %d", ip, c, count)
	}
	if o != over {
		t.Errorf("%q over quota, have: %t, want: %t", ip, o, over)
	}
	if e != expired {
		t.Errorf("%q expired, have: %t, want: %t", ip, e, expired)
	}
}

func TestQuota_Enabled(t *testing.T) {

	config.Quota.Window = 0
	config.Quota.Timeout = time.Second
	config.Quota.Stats = 0

	if comms.NewQuota().Enabled() {
		t.Errorf("Enabled() == true, want: false")
	}

	config.Quota.Window = time.Second
	if !comms.NewQuota().Enabled() {
		t.Errorf("Enabled() == false, want: true")
	}
}

func TestQuota_Query(t *testing.T) {

	config.Quota.Window = time.Second
	config.Quota.Timeout = time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota()

	// Test quota count, not over quota and valid
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		for x := 1; x <= maxQuota; x++ {
			qq.Quota(ip)
			verify(t, qq.Query, ip, x, UnderQuota, Valid)
		}
	}

	// Sleep so IP addresses expire
	time.Sleep(time.Second)

	// Test quota count, not over quota and expired
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		verify(t, qq.Query, ip, maxQuota, UnderQuota, Expired)
	}

	// Clear quota cache of expired IP addresses
	qq.CacheSweep()

	// Test IP address not in cache
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		verify(t, qq.Query, ip, IPNotCached, UnderQuota, Valid)
	}

	// Test quota count, over quota and valid
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		for x := 1; x <= maxQuota; x++ {
			qq.Quota(ip)
			verify(t, qq.Query, ip, x, UnderQuota, Valid)
		}
		// Push over quota
		qq.Quota(ip)
		verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)
	}

}

func TestQuota_CacheSweep(t *testing.T) {
	config.Quota.Window = time.Second
	config.Quota.Timeout = 2 * time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota()
	ip := "127.0.0.1"

	qq.Quota(ip)
	verify(t, qq.Query, ip, 1, UnderQuota, Valid)

	qq.CacheSweep()
	verify(t, qq.Query, ip, 1, UnderQuota, Valid)

	time.Sleep(config.Quota.Timeout)
	verify(t, qq.Query, ip, 1, UnderQuota, Expired)

	qq.CacheSweep()
	verify(t, qq.Query, ip, IPNotCached, false, false)
}

func TestQuota_CacheSweep_stats(t *testing.T) {
	config.Quota.Window = 10 * time.Second
	config.Quota.Timeout = 10 * time.Second
	config.Quota.Stats = 2 * time.Second

	// Intercept log
	config.Debug.LongLog = false
	log.SetFlags(0)
	out := &bytes.Buffer{}
	log.SetOutput(out)

	qq := comms.NewQuota()
	ip := "127.0.0.1"
	want := ""

	// Wait for stats due, then add IP address
	time.Sleep(config.Quota.Stats)
	out.Reset()
	qq.Quota(ip)
	verify(t, qq.Query, ip, 1, UnderQuota, Valid)
	qq.CacheSweep()
	want = "quota: 1 entries [0 evicted, 0 over quota, 1 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}

	// Wait for stats due, then take IP address over quota
	time.Sleep(config.Quota.Stats)
	out.Reset()
	qq.Quota(ip)
	qq.Quota(ip)
	qq.Quota(ip)
	qq.Quota(ip)
	verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)
	qq.CacheSweep()
	want = "quota: 1 entries [0 evicted, 1 over quota, 0 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}

	// Wait for IP address to go idle (timeout > stats due)
	time.Sleep(config.Quota.Timeout)
	out.Reset()
	qq.CacheSweep()
	verify(t, qq.Query, ip, IPNotCached, false, false)
	want = "quota: 1 entries [1 evicted, 0 over quota, 0 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}
}

func TestQuota_Quota_NoWindow(t *testing.T) {
	config.Quota.Window = 0
	config.Quota.Timeout = 0
	config.Quota.Stats = 0

	qq := comms.NewQuota()
	ip := "127.0.0.1"
	qq.Quota(ip)
	verify(t, qq.Query, ip, IPNotCached, false, false)
}

func TestQuota_Quota_Window(t *testing.T) {

	config.Quota.Window = 5 * time.Second
	config.Quota.Timeout = 5 * time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota()
	ip := "127.0.0.1"

	// Slowly use quota (takes 4 seconds), keeping under quota
	for x := 0; x < 4; x++ {
		qq.Quota(ip)
		time.Sleep(time.Second)
	}
	verify(t, qq.Query, ip, maxQuota, UnderQuota, Valid)

	// Delay long enough for two quota to expire, expired quota will be removed
	// when we add another.
	time.Sleep(2 * time.Second)
	qq.Quota(ip)
	verify(t, qq.Query, ip, 3, UnderQuota, Valid)

	// Let all but latest quota expire, expired quota will be removed
	// when we add another.
	time.Sleep(2 * time.Second)
	qq.Quota(ip)
	verify(t, qq.Query, ip, 2, UnderQuota, Valid)

	// Check IP address removed from cache
	time.Sleep(config.Quota.Timeout)
	qq.CacheSweep()
	verify(t, qq.Query, ip, IPNotCached, false, false)
}

func TestQuota_Quota_WithTimeout(t *testing.T) {

	config.Quota.Window = time.Second
	config.Quota.Timeout = 2 * time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota()
	ip := "127.0.0.1"

	// Go over quota
	for x := 1; x <= maxQuota+1; x++ {
		qq.Quota(ip)
	}
	verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)

	// Delay, but not expired
	time.Sleep(time.Second)
	verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)

	// Let expire
	time.Sleep(config.Quota.Timeout)
	verify(t, qq.Query, ip, maxQuota, OverQuota, Expired)

	// Check IP address removed from cache
	qq.CacheSweep()
	verify(t, qq.Query, ip, IPNotCached, false, false)
}

func TestQuota_Quota_WithoutTimeout(t *testing.T) {

	config.Quota.Window = 2 * time.Second
	config.Quota.Timeout = 0
	config.Quota.Stats = 0

	qq := comms.NewQuota()
	ip := "127.0.0.1"

	// Go over quota
	for x := 1; x <= maxQuota+1; x++ {
		qq.Quota(ip)
	}
	verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)

	// Delay, but not expired
	time.Sleep(time.Second)
	verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)

	// Let expire
	time.Sleep(config.Quota.Window)
	verify(t, qq.Query, ip, maxQuota, OverQuota, Expired)

	// Check IP address removed from cache
	qq.CacheSweep()
	verify(t, qq.Query, ip, IPNotCached, false, false)
}
