package extractor

import "github.com/alexpfx/go_search/downloader"

type New struct {
	Channel chan downloader.File
}

func (r New) Run() chan downloader.File {
	ch := r.Channel
	out := make(chan downloader.File)

	go func() {
		for f := range ch{
			extract(f)
			out <- f
		}
	}()

	return out
}

func extract(f downloader.File) {

}



