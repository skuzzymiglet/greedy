package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
)

type config struct { // TODO: add command line flags for every value here
	startPos       int
	wpm            int
	pauses         map[string]time.Duration // Determines the pause between words when the according string is found in the word
	normal, strong tcell.Style
	left           int // Start of text
}

func speedread(content []string, config config) (endPos int, err error) {
	// TODO: show "unfamiliar" words for longer
	// + Emojis
	// + "Weird" characters (symbols)
	// + Long words
	// Maybe we could pause based on an ngram database?

	// tcell stuff
	screen, err := tcell.NewScreen()
	if err != nil {
		return config.startPos, err
	}

	if err = screen.Init(); err != nil {
		return config.startPos, err
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
		word    int = config.startPos
		t       time.Duration
	)

	for word < len(content)-1 && word >= 0 {

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
		select { // TODO: listen for keys all the time
		case <-time.After(t):
		case key := <-keyChan:
			switch k := key.(type) {
			case *tcell.EventKey:
				switch k.Key() {
				case tcell.KeyLeft:
					if pausing {
						word--
					} else {
						word -= 2 // because we add 1 later
					}
				case tcell.KeyRight:
					word++
				}
				switch k.Rune() {
				case 'q':
					return word, nil
				case ' ':
					pausing = !pausing
				case ']':
					config.wpm += 10
				case '[':
					config.wpm -= 10
				case 'h':
					if pausing {
						word--
					} else {
						word -= 2 // because we add 1 later
					}
				case 'l':
					word++
				case '0':
					word = 0
				case '>':
					config.left++
				case '<':
					config.left--
				}
			}
		}
		screen.Clear()
		if !pausing {
			word++
		}
	}
	if word == len(content) {
		return 0, nil // If you reach the end, the position goes back to 0
	}
	return word, nil
}
