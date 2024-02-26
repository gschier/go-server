package main

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"time"
)

func main() {
	println("Worker started!")
	ticks := 0
	for {
		ticks++
		fmt.Printf("Worker tick %s!\n", aurora.Yellow(fmt.Sprintf("%d", ticks)))
		<-time.Tick(time.Millisecond * 7000)
	}
}
