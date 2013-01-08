package muxdaemons

import (
	"testing"
	"os/exec"
)

// check that we can load & run the test scripts
func TestCwd(t *testing.T) {
	out, err := exec.Command("testdata/1x0").Output()
	if err != nil {
		t.Error(err)
	}
	if string (out) != "1\n" {
		t.Error("bad output", out)
	}
}

/*
nope error
0x retire | 1x print
1x retire
x1 print 1
x1 print 2 | x2 print 1
x1 print 3 | x3 print 1
x1 print 4 | x2 print 2
x1 print 5
x1 retire
x2 print 3 | x3 print 2
x2 retire | x3 retire
*/

const eventType (
	output     int = iota
	retirement
)
type event struct {
	eventType eventType

}

func TestRun(t *testing.T) {
	outc := make(chan string)
	retiredc := make(chan *muxdaemons.Retired)
	go muxdaemons.Run([]string{"testdata/1x0", "testdata/5x1", "testdata/3x2", "testdata/2x3", "testdata/0x", "testdata/nope"}, outc, retiredc)
}
