package adapter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/amtoaer/epub-novel-builder/internal"
	"github.com/anaskhan96/soup"
	"github.com/carlmjohnson/requests"
	"golang.org/x/net/html"
)

type I69shu struct {
	titleMatcher *regexp.Regexp
}

func New69shu() *I69shu {
	return &I69shu{
		titleMatcher: regexp.MustCompile(`第.+?章`),
	}
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
	for start != nil {
		if start.Type == html.TextNode {
			data := strings.Trim(start.Data, "\n\r\t  ")
			if len(data) != 0 {
				paragraph = append(paragraph, data)
			}
		}
		start = start.NextSibling
	}
	title := chapter.Title
	if len(paragraph) >= 1 && strings.Contains(paragraph[len(paragraph)-1], "本章完") {
		paragraph = paragraph[:len(paragraph)-1]
	}
	if len(paragraph) >= 1 && i.titleMatcher.MatchString(paragraph[0]) {
		paragraph = paragraph[1:]
	}
	return internal.BookChapter{
		Title:   title,
		Content: internal.BuildContent(paragraph),
	}, nil
}

func (i *I69shu) Search(query string) []internal.BookInfo {
	body := bytes.NewBuffer([]byte{})
	query, _ = internal.Utf8StringToGbk(query)
	requests.URL("https://www.69shu.com/modules/article/search.php").BodyForm(url.Values{
		"searchkey": {query},
		"type":      {"all"},
	}).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8String(body)
	if err != nil {
		return nil
	}
	doc := internal.SafeSoup{Root: soup.HTMLParse(content)}
	if bookInfoItem, err := doc.SafeFind("div", "class", "newbox").SafeFind("ul").SafeFindAll("li"); err == nil {
		// 查询到多个结果，遍历结果列表
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
		return result
	}
	if bookInfo := doc.SafeFind("div", "class", "bookbox"); bookInfo.Error == nil {
		// 查询到单个结果，直接进入了详情页
		imgAttrs := bookInfo.Find("img").Attrs()
		title, coverPath := imgAttrs["alt"], imgAttrs["src"]
		author := bookInfo.Find("a", "target", "_blank").Attrs()["title"]
		bookDescriptionUrl := bookInfo.Find("a").Attrs()["href"]
		pathIndex, extensionIndex := strings.LastIndex(bookDescriptionUrl, "/"), strings.LastIndex(bookDescriptionUrl, ".")
		if pathIndex == -1 || extensionIndex == -1 || pathIndex >= extensionIndex {
			return nil
		}
		chapterUrl := fmt.Sprintf("https://www.69shu.com/%s/", bookDescriptionUrl[pathIndex+1:extensionIndex])
		return []internal.BookInfo{
			{
				Title:      title,
				Author:     author,
				CoverPath:  coverPath,
				ChapterUrl: chapterUrl,
			},
		}
	}
	// 无匹配
	return []internal.BookInfo{}
}

func (i *I69shu) Get(info internal.BookInfo) ([]internal.BookChapterInfo, error) {
	body := bytes.NewBuffer([]byte{})
	err := requests.URL(info.ChapterUrl).ToBytesBuffer(body).Fetch(context.Background())
	if err != nil {
		return nil, err
	}
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
