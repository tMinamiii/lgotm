package main

import (
	"flag"
	"log"
	"os"

	"github.com/tMinamiii/lgtm/lgtm"
)

func main() {
	mainText := flag.String("main", "L G T M", "main text")
	subText := flag.String("sub", "L o o k s   G o o d   T o   M e", "sub text")
	path := flag.String("i", "", "image path")
	textColor := flag.String("c", "white", "color 'white' or 'black'")
	serif := flag.Bool("serif", false, "serif font")
	gopher := flag.Bool("gopher", false, "embed gopher")
	flag.Parse()

	if *path == "" {
		log.Fatal("no image path")
		os.Exit(1)
	}

	if *textColor != "white" && *textColor != "black" {
		w := "white"
		textColor = &w
	}

	main := lgtm.NewText(*mainText, lgtm.FontSizeMain, lgtm.PointMain)
	sub := lgtm.NewText(*subText, lgtm.FontSizeSub, lgtm.PointSub)
	d := lgtm.NewTextDrawer(main, sub, *textColor, *serif, *gopher)

	if err := d.Draw(*path); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
