package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var wpm int = 400
var pauses map[string]time.Duration = map[string]time.Duration{
	".": time.Millisecond * time.Duration(500),
	"(": time.Millisecond * time.Duration(500),
	")": time.Millisecond * time.Duration(500),
	"-": time.Millisecond * time.Duration(300),
	",": time.Millisecond * time.Duration(300),
}

func main() {
	s := bufio.NewScanner(os.Stdin)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		w, _, err := terminal.GetSize(1)
		if err != nil {
			panic(err)
		}
		for i := 0; i < (w/2)-(len(s.Text())/2); i++ {
			fmt.Print(" ")
		}
		fmt.Print(s.Text())
		func() {
			for k, v := range pauses {
				if strings.Contains(s.Text(), k) {
					time.Sleep(v)
					return
				}
			}
			time.Sleep(time.Minute / time.Duration(wpm))
			return
		}()
		fmt.Print("\r")
		for i := 0; i < w; i++ {
			fmt.Print(" ")
		}
		fmt.Print("\r")
	}
}
