// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// NOTE: When adding quota for an IP adddress the fake time needs to advance.
// If not then the IP address will automatically ban itself when time appears
// not to change and the time of the first quota bucket == time of last bucket.

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

// Constants for use with verify to make tests more readable
const (
	UnderQuota  = false
	OverQuota   = true
	Valid       = false
	Expired     = true
	IPNotCached = -1
)

const localHost = "127.0.0.1"

// maxQuota is the maximum number of Quota an IP address can have within a
// config.Quota.Window period before going over quota.
var maxQuota = (&comms.Ring{}).Cap()

// fakeTime stores the fake 'current time' that is returned by the now
// function. It is set to a fixed time for reproducibility.
var fakeTime = time.Date(2019, time.May, 1, 13, 50, 55, 1, time.UTC)

// now implements a comms.TimeSource for a fake time for testing. Time can be
// manipulated by calling methods such as fakeTime.Add, allowing for  precise
// nano-second control of the testing time.
func now() time.Time { return fakeTime }

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

	if comms.NewQuota(now).Enabled() {
		t.Errorf("Enabled() == true, want: false")
	}

	config.Quota.Window = time.Second
	if !comms.NewQuota(now).Enabled() {
		t.Errorf("Enabled() == false, want: true")
	}
}

func TestQuota_Query(t *testing.T) {

	config.Quota.Window = time.Second
	config.Quota.Timeout = time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)

	// Test quota count, not over quota and valid
	for x := 1; x <= maxQuota; x++ {
		fakeTime = fakeTime.Add(time.Nanosecond)
		for off := 0; off < 256; off++ {
			ip := "127.0.0." + strconv.Itoa(off)
			qq.Quota(ip)
			verify(t, qq.Query, ip, x, UnderQuota, Valid)
		}
	}

	// Advance fake time so IP addresses expire
	fakeTime = fakeTime.Add(config.Quota.Window + time.Nanosecond)

	// Test quota count, not over quota and expired
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		verify(t, qq.Query, ip, maxQuota, UnderQuota, Expired)
	}

	// Clear quota cache of expired IP addresses
	qq.CacheSweep()

	// Check expired IP addresses cleared from cache
	for off := 0; off < 256; off++ {
		ip := "127.0.0." + strconv.Itoa(off)
		verify(t, qq.Query, ip, IPNotCached, UnderQuota, Valid)
	}

	// Test quota count, over quota and valid
	for x := 1; x <= maxQuota+1; x++ {
		fakeTime = fakeTime.Add(time.Nanosecond)
		for off := 0; off < 256; off++ {
			ip := "127.0.0." + strconv.Itoa(off)
			qq.Quota(ip)
			if x <= maxQuota {
				verify(t, qq.Query, ip, x, UnderQuota, Valid)
			} else {
				verify(t, qq.Query, ip, maxQuota, OverQuota, Valid)
			}
		}
	}

}

func TestQuota_CacheSweep(t *testing.T) {
	config.Quota.Window = time.Second
	config.Quota.Timeout = time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)

	qq.Quota(localHost)
	verify(t, qq.Query, localHost, 1, UnderQuota, Valid)

	qq.CacheSweep()
	verify(t, qq.Query, localHost, 1, UnderQuota, Valid)

	// Advance fake time to expire IP address
	fakeTime = fakeTime.Add(config.Quota.Window + time.Nanosecond)
	verify(t, qq.Query, localHost, 1, UnderQuota, Expired)

	// Check expired IP addresses cleared from cache
	qq.CacheSweep()
	verify(t, qq.Query, localHost, IPNotCached, false, false)
}

func TestQuota_CacheSweep_stats(t *testing.T) {
	config.Quota.Window = 3 * time.Second
	config.Quota.Timeout = 3 * time.Second
	config.Quota.Stats = time.Second

	// Intercept log so that we can inspect its content
	config.Debug.LongLog = false
	log.SetFlags(0)
	out := &bytes.Buffer{}
	log.SetOutput(out)

	qq := comms.NewQuota(now)
	want := ""

	// Advance fake time so stats are due (@ +1sec)
	fakeTime = fakeTime.Add(config.Quota.Stats + time.Nanosecond)

	// Add IP address
	out.Reset()
	qq.Quota(localHost)
	verify(t, qq.Query, localHost, 1, UnderQuota, Valid)
	qq.CacheSweep()
	want = "quota: 1 entries [0 evicted, 0 over quota, 1 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}

	// Advance fake time so stats are due (@ +2secs)
	fakeTime = fakeTime.Add(config.Quota.Stats + time.Nanosecond)

	// Take IP address over quota
	out.Reset()
	for x := 1; x <= maxQuota+1; x++ {
		fakeTime = fakeTime.Add(time.Nanosecond)
		qq.Quota(localHost)
	}
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)
	qq.CacheSweep()
	want = "quota: 1 entries [0 evicted, 1 over quota, 0 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}

	// Advance fake time so IP addresses go idle and stats due (@ +3secs)
	fakeTime = fakeTime.Add(config.Quota.Timeout + time.Nanosecond)

	out.Reset()
	qq.CacheSweep()
	verify(t, qq.Query, localHost, IPNotCached, false, false)
	want = "quota: 1 entries [1 evicted, 0 over quota, 0 recent]"
	if strings.TrimSpace(out.String()) != want {
		t.Errorf("CacheSweep stats:\nhave: %s\nwant: %s", out.String(), want)
	}
}

func TestQuota_Quota_NoWindow(t *testing.T) {
	config.Quota.Window = 0
	config.Quota.Timeout = 0
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)
	qq.Quota(localHost)
	verify(t, qq.Query, localHost, IPNotCached, false, false)
}

func TestQuota_Quota_Window(t *testing.T) {

	config.Quota.Window = 5 * time.Second
	config.Quota.Timeout = 0
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)

	// Slowly use up quota 1/second, keeping under quota (@ +maxQuota secs)
	for x := 1; x <= maxQuota; x++ {
		fakeTime = fakeTime.Add(time.Second)
		qq.Quota(localHost)
	}
	verify(t, qq.Query, localHost, maxQuota, UnderQuota, Valid)

	// Advance fake time so 2 quota expire. Expired quota will be removed when we
	// add another.
	fakeTime = fakeTime.Add(
		config.Quota.Window - 2*time.Second + time.Nanosecond,
	)
	qq.Quota(localHost)
	verify(t, qq.Query, localHost, maxQuota-2+1, UnderQuota, Valid)

	// Advance fake time so all but latest quota expire, expired quota will be
	// removed when we add another.
	fakeTime = fakeTime.Add(config.Quota.Window)
	qq.Quota(localHost)
	verify(t, qq.Query, localHost, 2, UnderQuota, Valid)

	// Advance fake time so last quota expires
	fakeTime = fakeTime.Add(config.Quota.Window + time.Nanosecond)

	// Check expired IP address removed from cache
	qq.CacheSweep()
	verify(t, qq.Query, localHost, IPNotCached, false, false)
}

func TestQuota_Quota_WithTimeout(t *testing.T) {

	config.Quota.Window = time.Second
	config.Quota.Timeout = 2 * time.Second
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)

	// Go over quota
	for x := 1; x <= maxQuota+1; x++ {
		fakeTime = fakeTime.Add(time.Nanosecond)
		qq.Quota(localHost)
	}
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)

	// Advance fake time so beyond Window but not timed out and expired yet
	fakeTime = fakeTime.Add(config.Quota.Window + time.Nanosecond)
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)

	// Advance fake time to beyond Timeout window and expire quota
	fakeTime = fakeTime.Add(
		config.Quota.Timeout - config.Quota.Window + time.Nanosecond,
	)
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Expired)

	// Check IP address removed from cache
	qq.CacheSweep()
	verify(t, qq.Query, localHost, IPNotCached, false, false)
}

func TestQuota_Quota_WithoutTimeout(t *testing.T) {

	config.Quota.Window = 5 * time.Second
	config.Quota.Timeout = 0
	config.Quota.Stats = 0

	qq := comms.NewQuota(now)

	// Slowly go over quota 1/second, (@ +maxQuota+1 secs)
	for x := 1; x <= maxQuota+1; x++ {
		fakeTime = fakeTime.Add(time.Second)
		qq.Quota(localHost)
	}
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)

	// Advance fake time to end of Window, but not over it
	fakeTime = fakeTime.Add(
		config.Quota.Window - time.Duration(maxQuota+1)*time.Second + time.Nanosecond,
	)
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)

	// Another connection attempt should extend waiting period
	qq.Quota(localHost)

	// Advance fake time by another Window, but not over it
	fakeTime = fakeTime.Add(config.Quota.Window)
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Valid)

	// Advance fake time past window so quota expires
	fakeTime = fakeTime.Add(time.Nanosecond)
	verify(t, qq.Query, localHost, maxQuota, OverQuota, Expired)

	// Check expired IP address removed from cache
	qq.CacheSweep()
	verify(t, qq.Query, localHost, IPNotCached, false, false)
}
