package adapter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amtoaer/epub-novel-builder/internal"
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

func (i *I69shu) Search(query string) []internal.BookInfo {
	body := bytes.NewBuffer([]byte{})
	query, _ = internal.Utf8StringToGbk(query)
	requests.URL("https://www.69shu.com/modules/article/search.php").BodyForm(url.Values{
		"searchkey": {query},
		"type":      {"all"},
	}).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8(body)
	if err != nil {
		return nil
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if bookInfoItem := doc.Find("div.newbox ul li"); len(bookInfoItem.Nodes) != 0 {
		result := make([]internal.BookInfo, 0, len(bookInfoItem.Nodes))
		bookInfoItem.Each(func(i int, s *goquery.Selection) {
			title := s.Find("h3").First().Children().Eq(1).Text()
			author := s.Find("div.labelbox").First().Children().First().Text()
			coverPath, _ := s.Find("img").Attr("data-src")
			descriptionUrl, _ := s.Find("a.imgbox").First().Attr("href")
			pathIndex, extensionIndex := strings.LastIndex(descriptionUrl, "/"), strings.LastIndex(descriptionUrl, ".")
			if pathIndex == -1 || extensionIndex == -1 || pathIndex >= extensionIndex {
				return
			}
			chapterUrl := fmt.Sprintf("https://www.69shu.com/%s/", descriptionUrl[pathIndex+1:extensionIndex])
			result = append(result, internal.BookInfo{
				Title:      title,
				Author:     author,
				CoverPath:  coverPath,
				ChapterUrl: chapterUrl,
			})
		})
		return result
	}
	if bookInfo := doc.Find("div.bookbox"); len(bookInfo.Nodes) != 0 {
		img := bookInfo.Find("img")
		title, _ := img.Attr("alt")
		coverPath, _ := img.Attr("src")
		author, _ := bookInfo.Find("a[target='_blank']").Attr("title")
		descriptionUrl, _ := bookInfo.Find("a").Attr("href")
		pathIndex, extensionIndex := strings.LastIndex(descriptionUrl, "/"), strings.LastIndex(descriptionUrl, ".")
		if pathIndex == -1 || extensionIndex == -1 || pathIndex >= extensionIndex {
			return nil
		}
		chapterUrl := fmt.Sprintf("https://www.69shu.com/%s/", descriptionUrl[pathIndex+1:extensionIndex])
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

func (i *I69shu) Get(info *internal.BookInfo) ([]internal.BookChapterInfo, error) {
	body := bytes.NewBuffer([]byte{})
	err := requests.URL(info.ChapterUrl).ToBytesBuffer(body).Fetch(context.Background())
	if err != nil {
		return nil, err
	}
	content, err := internal.GbkToUtf8(body)
	if err != nil {
		return nil, err
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(content))
	chapterList := doc.Find("div#catalog ul li")
	result := make([]internal.BookChapterInfo, 0, len(chapterList.Nodes))
	chapterList.Each(func(i int, s *goquery.Selection) {
		chapterInfo := s.Find("a")
		result = append(result, internal.BookChapterInfo{
			Title: chapterInfo.Text(),
			Url:   chapterInfo.AttrOr("href", ""),
		})
	})
	return result, nil
}

func (i *I69shu) Download(chapter internal.BookChapterInfo) (internal.BookChapter, error) {
	body := bytes.NewBuffer([]byte{})
	requests.URL(chapter.Url).ToBytesBuffer(body).Fetch(context.Background())
	content, err := internal.GbkToUtf8(body)
	if err != nil {
		return internal.BookChapter{}, err
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(content))
	chapterContent := doc.Find("div.txtnav")
	start := chapterContent.Nodes[0].FirstChild
	paragraph := []string{}
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
	if innerTitle := chapterContent.Find("h1.hide720").First().Text(); innerTitle != "" {
		title = innerTitle
	}
	if len(paragraph) >= 1 && strings.Contains(paragraph[len(paragraph)-1], "本章完") {
		paragraph = paragraph[:len(paragraph)-1]
	}
	if len(paragraph) >= 1 && i.titleMatcher.MatchString(paragraph[0]) {
		paragraph = paragraph[1:]
	}
	return internal.BookChapter{
		Title:   title,
		Content: internal.BuildContent(title, paragraph),
	}, nil
}
