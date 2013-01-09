package muxdaemons

import (
	"testing"
	"os/exec"
	"strings"
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
			strLine[i] = string(n)
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
		{out: {1: true}, ded: {1: true}},
		{out: {2: true, 3: true}, ded: {2: true, 3: true}},
	}
	/*
	events := []map[eventType] map[int] bool {
		{ded: {4: true}},
		{ded: {3: true}},
		{out: {0: true}},
		{out: {0: true, 1: true}},
		{out: {0: true, 2: true}},
		{out: {0: true, 1: true}},
		{out: {0: true}, ded: {0: true}},
		{out: {1: true, 2: true}, ded: {1: true, 2: true}},
	}
	*/
	outc := make(chan string)
	retiredc := make(chan *Retired)
	commands := []string{"testdata/1x0", "testdata/5x1", "testdata/3x2", "testdata/2x3", "testdata/0x", "testdata/nope"}
//	commands := []string{"testdata/5x1", "testdata/3x2", "testdata/2x3", "testdata/0x", "testdata/nope"}
	go Run(commands, outc, retiredc)
	var line [6]int
	for _, simultaneousEvents := range events {
		t.Logf("checking %v\n", simultaneousEvents)
		for len(simultaneousEvents) > 0 {
			select {
			case retiree := <-retiredc:
				t.Logf("got retiree %v\n", retiree)
				if simultaneousEvents[ded][retiree.Index] {
					delete(simultaneousEvents[ded], retiree.Index)
					if len(simultaneousEvents[ded]) == 0 {
						delete(simultaneousEvents, ded)
					}
				} else {
					t.Fatalf("wrongful death: got %d, expected one of %v", retiree.Index, simultaneousEvents[ded])
				}
			case output := <-outc:
				t.Logf("got output %s\n", output)
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
}
