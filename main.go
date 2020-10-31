package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/go-shiori/go-readability"
)

var wpm int = 400
var pauses map[string]time.Duration = map[string]time.Duration{
	".": time.Millisecond * time.Duration(500),
	"(": time.Millisecond * time.Duration(200),
	")": time.Millisecond * time.Duration(200),
	"-": time.Millisecond * time.Duration(300),
	",": time.Millisecond * time.Duration(300),
}

func main() {
	var scan *bufio.Scanner
	switch flag.NArg() {
	case 0:
		scan = bufio.NewScanner(os.Stdin)
	case 1:
		article, err := readability.FromURL("https://drewdevault.com/2019/10/12/how-to-fuck-up-releases.html", time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	scan := bufio.NewScanner(strings.NewReader(article.TextContent))
	scan.Split(bufio.ScanWords)

	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	if err = screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()
	screen.Clear()

	keyChan := make(chan tcell.Event, 0)
	go func() {
		for {
			keyChan <- screen.PollEvent()
		}
	}()
	var (
		w, h,
		start, // Leftmost character of each word
		bold int // Character to bolden
	)
	normal := tcell.StyleDefault
	strong := tcell.StyleDefault.Underline(true).Bold(true).Foreground(tcell.ColorRed)

	var tagline strings.Builder
	var word int                                          // current
	words := len(strings.Split(article.TextContent, " ")) // total
	for scan.Scan() {
		word++
		w, h = screen.Size() // TODO: only compute this on EventResize
		fmt.Fprintf(&tagline, "%d wpm (%d/%d) <", wpm, word, words)
		for i := 0; i < int((float64(word)/float64(words))*float64(w-len(tagline.String())-1)); i++ {
			fmt.Fprint(&tagline, "=")
		}
		for i, c := range tagline.String() {
			screen.SetContent(i, 0, c, []rune{}, tcell.StyleDefault)
		}
		screen.SetContent(w-1, 0, '>', []rune{}, tcell.StyleDefault)
		tagline.Reset()

		// start := (w / 2) - (len(scan.Text()) / 2)
		start = w / 2
		bold = w / 2
		for i, c := range scan.Text() {
			if start+i == bold {
				screen.SetContent(start+i, h/2, c, []rune{}, strong)
			} else {
				screen.SetContent(start+i, h/2, c, []rune{}, normal)
			}
		}
		screen.Show()
		var t time.Duration
		func() {
			for k, v := range pauses {
				if strings.Contains(scan.Text(), k) {
					t = v
					return
				}
			}
			t = time.Minute / time.Duration(wpm)
		}()
		select {
		case <-time.After(t):
		case key := <-keyChan:
			switch k := key.(type) {
			case *tcell.EventKey:
				switch k.Rune() {
				case 'q':
					return
				case ' ': // TODO: pause
				case ']':
					wpm += 10
				case '[':
					wpm -= 10
				}
			}
		}
		screen.Clear()
	}
}
