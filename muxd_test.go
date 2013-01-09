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

type eventType int
const (
	out eventType= iota
	ded
)
type event struct {
	eventType eventType
	command   string
	commandNo int
}

func TestRun(t *testing.T) {
	events := [][]event{
		{event{ded, "testdata/nope", 5}},
		{event{out, "testdata/1x0"}, event{ded, "testdata/0x"}},
		{event{ded, "testdata/1x0"}},
		{event{out, "testdata/5x1"}},
		{event{out, "testdata/5x1"}, event{out, "testdata/3x2"}},
		{event{out, "testdata/5x1"}, event{out, "testdata/2x3"}},
		{event{out, "testdata/5x1"}, event{out, "testdata/3x2"}},
		{event{out, "testdata/5x1"}},
		{event{ded, "testdata/5x1"}},
		{event{out, "testdata/3x2"}, event{out, "testdata/2x3"}},
		{event{ded, "testdata/3x2"}, event{ded, "testdata/2x3"}}
	}
	outc := make(chan string)
	retiredc := make(chan *muxdaemons.Retired)
	commands := []string{"testdata/1x0", "testdata/5x1", "testdata/3x2", "testdata/2x3", "testdata/0x", "testdata/nope"}
	go muxdaemons.Run(commands, outc, retiredc)
	var line [6]int
	for i, simultaneousEvents := range events {
		select {
		case retiree := <-retiredc:
			ok := false
			for j, event := range events {
				if event.eventType == ded && event.payload == commands[retiree] {
					ok = true

					break
				}
			}
			if !ok {
				t.Fatalf("wrongful death: got %s, expected one of %v", commands[retiree], events)
			}
		case out := <-outc:
			// I ain't gone be a slave to the data structure.
			// What I want to do is this: if I have simultaneous events, try out each possibility. Whichever one works out (if any),
			ok := false
			for j, event := range events {
				if event.eventType == out {
					line[event.commandNo] += 1
				}
			}
		}
	}
}
