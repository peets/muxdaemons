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
	go muxdaemons.Run(os.Args[1:], outc, retiredc)
	for {
		select {
		case muxed := <-outc:
			fmt.Println(muxed)
		case retired := <-retiredc:
			log.Printf("Retiring daemon %d (%s) because it exited with error %v\n", retired.Index, retired.Command, retired.Err)
		}
	}
}
