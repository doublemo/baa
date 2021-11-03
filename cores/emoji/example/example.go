package main

import (
	"flag"

	"github.com/doublemo/baa/cores/emoji"
)

func main() {
	emojiKeyword := flag.String("e", ":beer: Beer!!!", "emoji name")
	flag.Parse()

	emoji.Print(*emojiKeyword)
}
