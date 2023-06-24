package internal

type BookInfo struct {
	Title      string
	Author     string
	CoverPath  string
	CoverImage []byte
	ChapterUrl string
}

type BookChapterInfo struct {
	Title string
	Url   string
}

type BookChapter struct {
	Title   string
	Content string
}
