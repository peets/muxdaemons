// provides Run(), a function to run daemons and multiplex their output line by line
package muxdaemons

import (
	"os/exec"
	"strings"
	"bufio"
	"log"
)

const DEBUG bool = false

type MuxdCtl int
const (
	AllRetired MuxdCtl = iota
)

// Represents a retired daemon (a deamon that errored out)
type Retired struct {
	// the position of the command as given in the first argument to Run()
	Index int
	// the error that caused this daemon to retire (failed to start, broken pipe, EOF, etc)
	Err error
}

func collectOutput(i int, reader *bufio.Reader, out chan string, outSync chan int, retirements chan *Retired) {
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			out <- strings.TrimRight(line, "\n")
			outSync <- i
		}
		if err != nil {
			retirements <- &Retired{i, err}
			return
		}
	}
}

// Run daemons and multiplex their output line by line.
// 
// The given commands are launched in the background. The last line from each
// command's output is remembered. Whenever a command outputs a line, the
// remembered lines are concatenated; the resulting string is sent on the
// out channel.
//
// If a command fails to launch, or if its output stream is closed or broken,
// a Retired object will be sent on the retirements channel. When all commands are retired,
// AllRetired is sent on the ctl channel.
func Run(commands []string, out chan string, retirements chan *Retired, ctl chan MuxdCtl) {
	outSync := make(chan int)
	outs := make([]chan string, len(commands))
	retirementsProxy := make(chan *Retired, len(commands) + 1)

	// launch daemons
	for i, c := range commands {
		cmd := exec.Command(c)
		if DEBUG {
			log.Printf("%d (%s) execing\n", i, c)
		}
		cmdOut, err := cmd.StdoutPipe()
		if err != nil {
			retirementsProxy <- &Retired{i, err}
			continue
		}
		bufCmdOut := bufio.NewReader(cmdOut)
		if DEBUG {
			log.Printf("%d (%s) starting\n", i, c)
		}
		err = cmd.Start()
		if err != nil {
			retirementsProxy <- &Retired{i, err}
			continue
		}
		if DEBUG {
			log.Printf("%d (%s) started\n", i, c)
		}
		outs[i] = make(chan string, 1)
		go collectOutput(i, bufCmdOut, outs[i], outSync, retirementsProxy)
	}

	// mux output
	muxed := make([]string, len(commands))
	retired := 0
	for {
		select {
		case r := <-retirementsProxy:
			retirements <- r
			retired += 1
			if DEBUG {
				log.Printf("retired => %d; len(commands) => %d\n", retired, len(commands))
			}
			if retired == len(commands) {
				ctl <- AllRetired
				return
			}
		case i := <-outSync:
			if DEBUG {
				log.Printf("got sync %d\n", i)
			}
			muxed[i] = <-outs[i]
			if DEBUG {
				log.Printf("sending line '%v'\n", muxed)
			}
			out <- strings.Join(muxed, " ")
		}
	}
}
