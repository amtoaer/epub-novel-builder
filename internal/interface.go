package internal

type Adapter interface {
	Search(keyword string) ([]BookInfo, error)
	Get(book BookInfo) ([]BookChapterInfo, error)
	Download(chapter BookChapterInfo) (BookChapter, error)
}
