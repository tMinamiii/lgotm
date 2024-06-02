package main

import (
	"flag"
	"log"
	"os"

	"github.com/tMinamiii/lgtm/drawer"
	"github.com/tMinamiii/lgtm/object"
)

func main() {
	path := flag.String("i", "", "image file path")
	color := flag.String("c", "white", "color 'white' or 'black'")
	fontName := flag.String("f", "sans", "sans, serif, line")
	gopher := flag.Bool("gopher", false, "embed gopher")
	flag.Parse()

	if *path == "" {
		log.Fatal("no image path")
		os.Exit(1)
	}

	if *gopher {
		d := drawer.NewGopherDrawer()
		if err := d.Draw(*path); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		return
	}

	textColor := object.TextColorWhite
	if *color == "black" {
		textColor = object.TextColorBlack
	}

	font := getFont(*fontName)
	main := object.NewText(object.DefaultMainText, font, object.MessageTypeMain, textColor)
	sub := object.NewText(object.DefaultSubText, font, object.MessageTypeSub, textColor)

	d := drawer.NewTextDrawer(main, sub)
	if err := d.Draw(*path); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func getFont(fontName string) object.Font {
	switch fontName {
	case "serif":
		return object.NotoSerifJP
	case "line":
		return object.LINESeedJP
	case "sans":
		return object.NotoSansJP
	default:
		return object.NotoSansJP
	}
}
