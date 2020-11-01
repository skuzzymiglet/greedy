package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/go-shiori/go-readability"
)

type config struct { // TODO: add command line flags for every value here
	startPos       int
	wpm            int
	pauses         map[string]time.Duration // Determines the pause between words when the according string is found in the word
	normal, strong tcell.Style
	left           int // Start of text
}

// Find last stored position
// Returns 0, nil if none is found and no error occurs
func lookupPos(contentHash [sha256.Size]byte) (int, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return 0, err
	}
	pos, err := ioutil.ReadFile(filepath.Join(cacheDir, "greedy", hex.EncodeToString(contentHash[:])))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.Atoi(string(pos))
}

func writePos(contentHash [sha256.Size]byte, pos int) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(cacheDir, "greedy"), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(
		filepath.Join(cacheDir, "greedy", hex.EncodeToString(contentHash[:])),
		[]byte(strconv.Itoa(pos)),
		0777,
	)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var (
		wpm       int
		savePos   bool
		resumePos bool
	)
	flag.IntVar(&wpm, "w", 400, "words per minute")
	flag.BoolVar(&savePos, "p", true, "save position")
	flag.BoolVar(&resumePos, "r", true, "try to resume at saved position")
	flag.Parse()
	// Get content
	var content []string
	var contentHash [sha256.Size]byte
	switch flag.NArg() {
	case 0:
		fmt.Fprintf(os.Stderr, "%s: reading from stdin...\n", filepath.Base(os.Args[0]))
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalln("error reading stdin:", err)
		}
		contentHash = sha256.Sum256(b)
		content = strings.Fields(string(b))
	case 1:
		article, err := readability.FromURL(flag.Arg(0), time.Second*3)
		if err != nil {
			log.Fatalln("error extracting article text:", err)
		}
		contentHash = sha256.Sum256([]byte(article.TextContent))
		content = strings.Fields(article.TextContent)
		// TODO: handle images, code blocks, links, footnotes and other web stuff
	}
	conf := config{
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
	}
	if resumePos {
		p, err := lookupPos(contentHash)
		if err != nil {
			log.Fatalln("error finding last position:", err)
		}
		conf.startPos = p
	}
	end, err := speedread(content, conf)
	if savePos {
		err = writePos(contentHash, end)
		if err != nil {
			log.Fatalln("Error saving position:", err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func speedread(content []string, config config) (endPos int, err error) {
	// TODO: resume at a position
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
	return word, nil
}
