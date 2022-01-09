// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package quota_test

import (
	"strings"
	"testing"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/quota"
)

// fakeTime stores the fake 'current time' that is returned by the now
// function. It is set to a fixed time for reproducibility.
var fakeTime = time.Date(2019, time.May, 1, 13, 50, 55, 1, time.UTC)

// now implements a comms.TimeSource for a fake time for testing. Time can be
// manipulated by calling methods such as fakeTime.Add, allowing for precise
// nano-second control of the testing time.
func now() time.Time { return fakeTime }

var cfg = `// Test configuration
  Quota.Slots:  2
  Quota.Window: 10s
	Debug.Quota:	false
`

func good(t *testing.T, slots uint64, adv time.Duration) {
	t.Helper()
	fakeTime = fakeTime.Add(adv)
	if !quota.Accept("127.0.0.1") {
		t.Errorf("should be good, was over")
	}
	if have := quota.CacheBits("127.0.0.1"); have != slots {
		t.Errorf("cache bits incorrect,\nhave: %064b\nwant: %064b", have, slots)
	}
}

func over(t *testing.T, slots uint64, adv time.Duration) {
	t.Helper()
	fakeTime = fakeTime.Add(adv)
	if quota.Accept("127.0.0.1") {
		t.Errorf("should be over, was good")
	}
	if have := quota.CacheBits("127.0.0.1"); have != slots {
		t.Errorf("cache bits incorrect,\nhave: %064b\nwant: %064b", have, slots)
	}
}

func setup(t *testing.T) {
	t.Helper()
	b := strings.NewReader(cfg)
	c := config.Config{}
	c, _ = c.Read(b)
	quota.Config(c, now)
}

func TestQuota_ggg(t *testing.T) {
	setup(t)
	good(t, 0b00000000000000000000000000000001, 0)
	good(t, 0b00000000000000000000000000000101, 5000*time.Millisecond)
	good(t, 0b00000000000000000000000000010101, 5000*time.Millisecond)
}

func TestQuota_ggo(t *testing.T) {
	setup(t)
	good(t, 0b00000000000000000000000000000001, 0)
	good(t, 0b00000000000000000000000000000011, 4900*time.Millisecond)
	over(t, 0b00000000000000000000000000000111, 4900*time.Millisecond)
}

func TestQuota_ggoSmall(t *testing.T) {
	setup(t)
	good(t, 0b00000000000000000000000000000001, 0)
	good(t, 0b00000000000000000000000000000011, 4900*time.Millisecond)
	over(t, 0b00000000000000000000000000000111, 5000*time.Millisecond)
}

func TestQuota_ggoo(t *testing.T) {
	setup(t)
	good(t, 0b00000000000000000000000000000001, 0)
	good(t, 0b00000000000000000000000000000011, 4900*time.Millisecond)
	over(t, 0b00000000000000000000000000000111, 4900*time.Millisecond)
	over(t, 0b00000000000000000000000000001111, 4900*time.Millisecond)
}

func TestQuota_ggog(t *testing.T) {
	setup(t)
	good(t, 0b00000000000000000000000000000001, 0)
	good(t, 0b00000000000000000000000000000011, 4900*time.Millisecond)
	over(t, 0b00000000000000000000000000000111, 4900*time.Millisecond)
	good(t, 0b00000000000000000000000000011101, 5000*time.Millisecond)
}

func TestQuota_gggOverflow(t *testing.T) {
	setup(t)
	good(t, 0b0000000000000000000000000000000000000000000000000000000000000001, 0)
	good(t, 0b1000000000000000000000000000000000000000000000000000000000000001,
		310000*time.Millisecond)
	good(t, 0b0000000000000000000000000000000000000000000000000000000000000011,
		1000*time.Millisecond)
}
