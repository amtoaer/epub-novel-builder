package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/amtoaer/epub-novel-builder/internal"
	"github.com/amtoaer/epub-novel-builder/internal/adapter"
	"github.com/bmaupin/go-epub"
)

func main() {
	t := adapter.New69shu()
	books := t.Search("将夜")
	if len(books) == 0 {
		panic("no book found")
	}
	book := books[0]
	e := epub.NewEpub(book.Title)
	cssPath, err := e.AddCSS("./internal/custom.css", "custom.css")
	if err != nil {
		panic(err)
	}
	e.SetAuthor(book.Author)
	if book.CoverPath != "" {
		path, err := e.AddImage(book.CoverPath, "cover.png")
		if err != nil {
			panic(err)
		}
		e.SetCover(path, "")
	}
	chapterInfos, _ := t.Get(&book)
	for _, chapterInfo := range chapterInfos {
		chapter, _ := t.Download(chapterInfo)
		e.AddSection(chapter.Content, chapter.Title, "", cssPath)
		fmt.Printf("successfully download %s!\n", chapter.Title)
		time.Sleep(internal.TIME_TO_SLEEP * time.Millisecond)
	}
	println("saving to " + path.Join(internal.OUTPUT_DIR, book.Title+".epub") + " ...")
	os.MkdirAll(internal.OUTPUT_DIR, os.FileMode(0755))
	e.Write(path.Join(internal.OUTPUT_DIR, book.Title+".epub"))
	println("success")
}
