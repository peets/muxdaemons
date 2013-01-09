// run given daemons and multiplex their output line by line
package main

import (
	"fmt"
	"os"
	"log"
	"ppplog.net/muxdaemons"
)

func main() {
	outc := make(chan string)
	retiredc := make(chan *muxdaemons.Retired)
	ctl := make(chan muxdaemons.MuxdCtl)
	commands := os.Args[1:]
	go muxdaemons.Run(commands, outc, retiredc, ctl)
	for {
		select {
		case muxed := <-outc:
			fmt.Println(muxed)
		case retired := <-retiredc:
			log.Printf("Retiring daemon %d (%s) because it exited with error %v\n", retired.Index, commands[retired.Index], retired.Err)
		case ctlEvent := <-ctl:
			switch ctlEvent {
			case muxdaemons.AllRetired:
				return
			}
		}
	}
}
