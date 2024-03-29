package muxdaemons

import (
	"testing"
	"os/exec"
	"strings"
	"fmt"
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

type eventType int
const (
	out eventType = iota
	ded
)

func checkLine(line [6]int, ref string) bool {
	strLine := make([]string, 6)
	for i, n := range line {
		if n == 0 {
			strLine[i] = ""
		} else {
			strLine[i] = fmt.Sprintf("%d", n)
		}
	}
	return strings.Join(strLine, " ") == ref
}

func TestRun(t *testing.T) {
	events := []map[eventType] map[int] bool {
		{ded: {0: true, 4: true, 5: true}, out: {0: true}},
		{out: {1: true}},
		{out: {1: true, 2: true}},
		{out: {1: true, 3: true}},
		{out: {1: true, 2: true}},
		{out: {1: true}},
		{ded: {1: true}},
		{out: {2: true, 3: true}, ded: {2: true, 3: true}},
	}
	outc := make(chan string)
	retiredc := make(chan *Retired)
	ctl := make(chan MuxdCtl)
	commands := []string{"testdata/1x0", "testdata/5x1", "testdata/3x2", "testdata/2x3", "testdata/0x", "testdata/nope"}
	go Run(commands, outc, retiredc, ctl)
	var line [6]int
	for _, simultaneousEvents := range events {
		if DEBUG {
			t.Logf("checking %v\n", simultaneousEvents)
		}
		for len(simultaneousEvents) > 0 {
			select {
			case retiree := <-retiredc:
				if DEBUG {
					t.Logf("got retiree %v\n", retiree)
				}
				if simultaneousEvents[ded][retiree.Index] {
					delete(simultaneousEvents[ded], retiree.Index)
					if len(simultaneousEvents[ded]) == 0 {
						delete(simultaneousEvents, ded)
					}
				} else {
					t.Fatalf("wrongful death: got %d, expected one of %v", retiree.Index, simultaneousEvents[ded])
				}
			case output := <-outc:
				if DEBUG {
					t.Logf("got output '%s'\n", output)
				}
				ok := false
				for c, _ := range simultaneousEvents[out] {
					tentativeLine := line
					tentativeLine[c] += 1
					if(checkLine(tentativeLine, output)) {
						line[c] += 1
						delete(simultaneousEvents[out], c)
						if len(simultaneousEvents[out]) == 0 {
							delete(simultaneousEvents, out)
						}
						ok = true
						break
					}
				}
				if !ok {
					t.Fatalf("output '%s' cannot be reached from line '%v' with any of these events: %v", output, line, simultaneousEvents[out])
				}
			}
		}
	}
	for {
		select {
		case ev := <-ctl:
			if ev != AllRetired {
				t.Fatalf("got a ctl signal other than AllRetired\n")
			}
			return
		case output := <-outc:
			t.Fatalf("leftover output: '%s'\n", output)
		case retiree := <-retiredc:
			t.Fatalf("leftover retiree: '%v'\n", retiree)
		}
	}
}
