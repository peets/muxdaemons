// provides Run(), a function to run daemons and multiplex their output line by line
package muxdaemons

import (
	"os/exec"
	"strings"
	"bufio"
	"log"
)

// Represents a retired daemon (a deamon that errored out)
type Retired struct {
	// the position of the command as given in the first argument to Run()
	Index int
	// the command string as given in the first argument to Run()
	Command string
	// the error that caused this daemon to retire (failed to start, broken pipe, EOF, etc)
	Err error
}

type daemon struct {
	out		chan string
	command	string
	err		error
}

func teeLine(r *bufio.Reader, errc chan error) (string) {
	str, err := r.ReadString('\n')
	if err != nil {
		errc <- err
	}
	return str
}

// Run daemons and multiplex their output line by line.
// 
// The given commands are launched in the background. The last line from each
// command's output is remembered. Whenever a command outputs a line, the
// remembered lines are concatenated; the resulting string is sent on the
// out channel.
//
// If a command fails to launch, or if its output stream is closed or broken,
// a Retired object will be sent on the retirements channel.
func Run(commands []string, out chan string, retirements chan *Retired) {
	// initialize daemons
	daemons := make([]*daemon, len(commands))
	for i, c := range commands {
		daemons[i] = &daemon{make(chan string, 1), c, nil}
	}

	// launch daemons
	errc := make(chan int)
	sync := make(chan int)
	for i, d := range daemons {
		go func(i int, d *daemon) {
			lines := make(chan string, 1)
			myErrc := make(chan error, 1)
			line := ""

			log.Printf("%d (%s) execing\n", i, d.command)
			cmd := exec.Command(d.command)
			cmdOut, err := cmd.StdoutPipe()
			if err != nil {
				d.err = err
				errc <- i
				return
			}
			bufCmdOut := bufio.NewReader(cmdOut)
			log.Printf("%d (%s) starting\n", i, d.command)
			err = cmd.Start()
			if err != nil {
				d.err = err
				errc <- i
				return
			}
			log.Printf("%d (%s) started\n", i, d.command)
			for {
				select {
				case lines <- teeLine(bufCmdOut, myErrc):
					log.Printf("%d (%s) teeLine out\n", i, d.command)
				case line = <-lines:
					log.Printf("%d (%s) line consumed\n", i, d.command)
					d.out <- line
					sync <- i
				case err := <-myErrc:
					log.Printf("%d (%s) error consumed\n", i, d.command)
					d.err = err
					errc <- i
					return
				}
			}
		}(i, d)
	}

	// mux output as it arrives
	lines := make([]string, len(daemons))
	for {
		select {
		case i := <-errc:
			log.Printf("sending retirement %d\n", i)
			retirements <- &Retired{i, daemons[i].command, daemons[i].err}
		case i := <-sync:
			log.Printf("got sync %d\n", i)
			lines[i] = strings.TrimRight(<-daemons[i].out, "\n")
			log.Printf("sending line '%v'\n", lines)
			out <- strings.Join(lines, " ")
		}
	}
}
