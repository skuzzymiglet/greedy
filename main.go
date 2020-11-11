package main

import (
	"crypto/sha256"
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

var defaultConfig = config{
	wpm:    400,
	strong: tcell.StyleDefault.Bold(true).Foreground(tcell.ColorRed),
	// normal: tcell.StyleDefault.Reverse(true),
	pauses: map[string]time.Duration{
		".": time.Millisecond * time.Duration(500),
		"(": time.Millisecond * time.Duration(200),
		")": time.Millisecond * time.Duration(200),
		"-": time.Millisecond * time.Duration(300),
		",": time.Millisecond * time.Duration(300),
	},
	pauseEnd: true, pauseStart: true,
	left: 10,
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
	// TODO: stream mode where input is read gradually. Cos stdin
	var content []string
	var contentHash [sha256.Size]byte
	var title string
	switch flag.NArg() {
	case 0:
		title = "stdin"
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
		title = article.Title
		contentHash = sha256.Sum256([]byte(article.TextContent))
		content = strings.Fields(article.TextContent)
		// TODO: handle images, code blocks, links, footnotes and other web stuff
	}
	// TODO: cannot resume a file if at end
	conf := defaultConfig
	conf.wpm = wpm
	if resumePos {
		p, err := lookupPos(contentHash)
		if err != nil {
			log.Fatalln("error finding last position:", err)
		}
		log.Printf("Resuming at word %d\n", p)
		conf.startPos = p
	}
	end, err := speedread(content, conf, title)
	if savePos {
		log.Println("Saving position...")
		err = writePos(contentHash, end)
		if err != nil {
			log.Fatalln("Error saving position:", err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
