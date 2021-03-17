package downloader

import (
	"testing"
)

func TestDownload_ListTopFileSizes(t *testing.T) {
	baseUrl := "http://www.gutenberg.org/robot/"
	page := "harvest"
	c := Crawler{
		Start: baseUrl + page,
		DownloadIf: func(fileUrl string, sizeMb float64) bool {
			return sizeMb > 12
		},
	}
	ch := c.Run()

	downloader := Downloader{
		TargetDir: "/opt/datasets/large",
		Channel:   ch,
	}
	downloader.Run()

}

func up(low *float64, high *float64, mag float64) {
	*low *= mag
	*high *= mag

	//fmt.Printf("low: %2.2f high: %2.2f \n", *low, *high)
}
