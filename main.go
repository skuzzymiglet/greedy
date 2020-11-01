package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/go-shiori/go-readability"
)

type config struct { // TODO: add command line flags for every value here
	wpm            int
	pauses         map[string]time.Duration // Determines the pause between words when the according string is found in the word
	normal, strong tcell.Style
	left           int // Start of text
}

func main() {
	var wpm int
	flag.IntVar(&wpm, "w", 400, "words per minute")
	flag.Parse()
	// Get content
	var content []string
	switch flag.NArg() {
	case 0:
		fmt.Fprintf(os.Stderr, "%s: reading from stdin...\n", filepath.Base(os.Args[0]))
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalln("error reading stdin:", err)
		}
		content = strings.Fields(string(b))
	case 1:
		article, err := readability.FromURL(flag.Arg(0), time.Second*3)
		if err != nil {
			log.Fatalln("error extracting article text:", err)
		}
		content = strings.Fields(article.TextContent)
		// TODO: handle images, code blocks, links, footnotes and other web stuff
	}
	err := speedread(content, config{
		wpm:    wpm,
		strong: tcell.StyleDefault.Bold(true).Foreground(tcell.ColorRed),
		// normal: tcell.StyleDefault.Reverse(true),
		pauses: map[string]time.Duration{
			".": time.Millisecond * time.Duration(500),
			"(": time.Millisecond * time.Duration(200),
			")": time.Millisecond * time.Duration(200),
			"-": time.Millisecond * time.Duration(300),
			",": time.Millisecond * time.Duration(300),
		},
		left: 10,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func speedread(content []string, config config) error { // TODO: turn config parameters into a struct
	// tcell stuff
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err = screen.Init(); err != nil {
		return err
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
		pausing bool
		w, h,
		bold int // Character to bolden
		tagline strings.Builder
		word    int // current word
		t       time.Duration
	)

	for word < len(content)-1 && word >= 0 {
		if !pausing {
			word++
		}

		w, h = screen.Size() // TODO: only compute this on EventResize
		read := float64(word) / float64(len(content))

		// tagline
		fmt.Fprintf(&tagline, "%d wpm (%d/%d) <", config.wpm, word, len(content))
		taglineWidth := w - len(tagline.String()) - 1

		for i := 0; i < int(read*float64(taglineWidth)); i++ {
			fmt.Fprint(&tagline, "=")
		}

		for i, c := range tagline.String() {
			screen.SetContent(i, 0, c, []rune{}, tcell.StyleDefault)
		}
		screen.SetContent(w-1, 0, '>', []rune{}, tcell.StyleDefault)
		tagline.Reset()

		// word
		bold = config.left
		for i, c := range content[word] {
			// if !unicode.IsGraphic(c) {
			// 	panic("non-graphic char")
			// }
			if config.left+i == bold {
				screen.SetContent(config.left+i, h/2, c, []rune{}, config.strong)
			} else {
				screen.SetContent(config.left+i, h/2, c, []rune{}, config.normal)
			}
		}
		screen.Show()

		// determine how long to wait
		t = func() time.Duration {
			for k, v := range config.pauses {
				if strings.Contains(content[word], k) {
					return v
				}
			}
			return time.Minute / time.Duration(config.wpm)
		}()
		select {
		case <-time.After(t):
		case key := <-keyChan:
			switch k := key.(type) {
			case *tcell.EventKey:
				switch k.Rune() {
				case 'q':
					return nil
				case ' ':
					pausing = !pausing // TODO: handle events, update tagline while pausing
				case ']':
					config.wpm += 10
				case '[':
					config.wpm -= 10
				case 'h':
					word-- // BUG: doesn't work
				case 'l':
					word++
				case '>':
					config.left++
				case '<':
					config.left--
				}
			}
		}
		screen.Clear()
	}
	return nil
}
