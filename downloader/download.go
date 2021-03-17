package downloader

import (
	"github.com/PuerkitoBio/goquery"
	_ "github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Downloader struct {
	TargetDir string
	Channel   chan File
}

func (d Downloader) Run() chan File {
	ch := make(chan File, 4)
	go func() {
		for file := range d.Channel {
			res, ok := get(file.url)
			if !ok {
				continue
			}

			s := strings.Split(file.url, "/")
			targetPath := filepath.Join(d.TargetDir, s[len(s)-1])
			log.Printf("downloading %s to %s. size: %2.2fMB\n", file.url, targetPath, file.length)
			downloaded := downloadTo(res, targetPath)
			if downloaded {
				ch <- File{
					url:        file.url,
					length:     file.length,
					targetPath: file.targetPath,
				}
			}
		}
	}()

	return ch
}

type Crawler struct {
	Start      string
	DownloadIf func(fileUrl string, sizeMb float64) bool
}
type File struct {
	url        string
	length     float64
	targetPath string
}

func (c Crawler) Run() chan File {
	files := make(chan File, 8)

	go func() {
		next, err := url.Parse(c.Start)
		if err != nil {
			return
		}

		for {

			log.Println("next: ", next)

			res, ok := get(next.String())
			if !ok {
				return
			}

			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
				return
			}

			var shouldContinue bool
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if !exists {
					return
				}

				if strings.HasSuffix(href, ".zip") {
					shouldGet, length := c.checkDownload(href)
					if shouldGet {
						files <- File{url: href, length: length}
					}
				} else if strings.Contains(href, "offset=") {
					next, err = next.Parse(href)
					if err != nil {
						shouldContinue = false
					}
					shouldContinue = true
				}
			})
			res.Body.Close()
			if !shouldContinue {
				close(files)
				break
			}
		}
	}()

	return files

}

func head(url string) (*http.Response, bool) {
	res, err := http.Head(url)
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}
	if res.StatusCode != 200 {
		return nil, false
	}
	return res, true
}

func get(url string) (*http.Response, bool) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, false
	}
	if resp.StatusCode != 200 {
		return nil, false
	}

	return resp, true
}

func (c Crawler) checkDownload(fileUrl string) (bool, float64) {
	res, ok := head(fileUrl)
	if !ok {
		return false, 0
	}

	mb := toMB(res.ContentLength)

	return c.DownloadIf(fileUrl, mb), mb
}

func downloadTo(res *http.Response, target string) bool {
	defer res.Body.Close()
	file, err := os.Create(target)
	if err != nil {
		log.Fatal(file)
		return false
	}

	defer file.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func toMB(length int64) float64 {
	return float64(length) / (1024 * 1024)
}
