package adapter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/amtoaer/epub-novel-builder/internal"
	"github.com/anaskhan96/soup"
	"github.com/carlmjohnson/requests"
	"golang.org/x/net/html"
)

type I69shu struct {
}

func (i *I69shu) Download(chapter internal.BookChapterInfo) (internal.BookChapter, error) {
	body := bytes.NewBuffer([]byte{})
	requests.URL(chapter.Url).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8String(body)
	if err != nil {
		return internal.BookChapter{}, err
	}
	doc := soup.HTMLParse(content)
	chapterContent := doc.Find("div", "class", "txtnav")
	start := chapterContent.Pointer.FirstChild
	var paragraph []string
	var buf bytes.Buffer
	for start != nil {
		if start.Type == html.TextNode {
			data := strings.Trim(start.Data, "\n\r\t  ")
			if len(data) != 0 {
				buf.WriteString(internal.PARAGRAPH_PREFIX)
				buf.WriteString(data)
				paragraph = append(paragraph, buf.String())
				buf.Reset()
			}
		}
		start = start.NextSibling
	}
	title := chapter.Title
	if len(paragraph) >= 2 {
		// 第一行是标题，最后一行是“（本章完）”，需要跳过
		title = strings.TrimPrefix(paragraph[0], internal.PARAGRAPH_PREFIX)
		paragraph = paragraph[1 : len(paragraph)-1]
	}
	return internal.BookChapter{
		Title:   title,
		Content: internal.BuildContent(paragraph),
	}, nil
}

func (i *I69shu) Search(query string) ([]internal.BookInfo, error) {
	body := bytes.NewBuffer([]byte{})
	query, _ = internal.Utf8StringToGbk(query)
	requests.URL("https://www.69shu.com/modules/article/search.php").BodyForm(url.Values{
		"searchkey": {query},
		"type":      {"all"},
	}).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8String(body)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(content)
	bookInfoItem := doc.Find("div", "class", "newbox").Find("ul").FindAll("li")
	result := make([]internal.BookInfo, 0, len(bookInfoItem))
	for _, bookInfo := range bookInfoItem {
		infoBlock := bookInfo.Find("div", "class", "newnav")
		bookDescriptionUrl := infoBlock.Find("a", "class", "imgbox").Attrs()["href"]
		pathIndex, extensionIndex := strings.LastIndex(bookDescriptionUrl, "/"), strings.LastIndex(bookDescriptionUrl, ".")
		if pathIndex == -1 || extensionIndex == -1 || pathIndex >= extensionIndex {
			continue
		}
		chapterUrl := fmt.Sprintf("https://www.69shu.com/%s/", bookDescriptionUrl[pathIndex+1:extensionIndex])
		tmp := internal.BookInfo{
			Title:      infoBlock.Find("h3").Children()[1].Text(),
			Author:     infoBlock.Find("div", "class", "labelbox").Find("label").Text(),
			CoverPath:  bookInfo.Find("img").Attrs()["data-src"],
			ChapterUrl: chapterUrl,
		}
		result = append(result, tmp)
	}
	return result, nil
}

func (i *I69shu) Get(info internal.BookInfo) ([]internal.BookChapterInfo, error) {
	body := bytes.NewBuffer([]byte{})
	requests.URL(info.ChapterUrl).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8String(body)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(content)
	chapterList := doc.Find("div", "id", "catalog").Find("ul").FindAll("li")
	result := make([]internal.BookChapterInfo, 0, len(chapterList))
	for _, chapter := range chapterList {
		chapterInfo := chapter.Find("a")
		tmp := internal.BookChapterInfo{
			Title: chapterInfo.Text(),
			Url:   chapterInfo.Attrs()["href"],
		}
		result = append(result, tmp)
	}
	return result, nil
}
