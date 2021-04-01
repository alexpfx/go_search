package main

import (
	"fmt"
	downloader2 "github.com/alexpfx/go_search/downloader"
	"github.com/alexpfx/go_search/extractor"
	"github.com/alexpfx/go_search/search"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"time"
)

var clCyan = color.New(color.FgHiCyan)
var clBlue = color.New(color.FgHiBlue)
var plain bool

var clResult = clBlue.SprintFunc()

var clPathResult = clCyan.SprintFunc()

func main() {
	app := &cli.App{
		HideHelpCommand: true,
		Name:            "go_search",
		Usage:           "buscador de texto em arquivos",
		ArgsUsage:       "<query_string>",

		Commands: []*cli.Command{
			{
				Name: "gen",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "extract",
						Aliases: []string{"x"},
						Usage:   "indica que o arquivo obtido via download deve ser extraído",
						Value:   true,
					},
					&cli.StringFlag{Usage: "diretório onde os arquivos serão gravados",
						Aliases: []string{"d"},
						Name:    "targetDir",
						Value:   "./",
					},
					&cli.IntFlag{

						Name:    "size",
						Usage:   "só realiza o download do arquivos com tamanho maior que <size>MB",
						Aliases: []string{"s"},
						Value:   5,
					}},
				Usage: "Obtém arquivos texto do site http://www.gutenberg.org para fins de teste",
				Action: func(c *cli.Context) error {
					dir := c.String("targetDir")
					size := c.Int("size")

					baseUrl := "http://www.gutenberg.org/robot/harvest"
					crawler := downloader2.Crawler{
						Start: baseUrl,
						DownloadIf: func(fileUrl string, sizeMb float64) bool {
							return sizeMb > float64(size)
						},
					}
					ch := crawler.Run()

					downloader := downloader2.Downloader{
						TargetDir: dir,
						Channel:   ch,
					}

					extract := c.Bool("extract")

					downloadCh := downloader.Run()

					if extract {
						ext := extractor.New{Channel: downloadCh}
						extCh := ext.Run()

						for f := range extCh {
							log.Println("extraido: ", f)
						}

					}

					return nil

				},
			},
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Usage:   "diretório raíz de onde a busca deve partir",
				Value:   "./",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "inclui na busca diretórios invisíveis",
				Value: false,
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage: `busca o texto apenas que contenham uma parte da string no nome. 
					exemplos de uso: 
						-i txt							
						-i txt -i java -i main.go					
				`,
			},
			&cli.BoolFlag{
				Name:    "plain",
				Aliases: []string{"p"},
				Value:   false,
				Usage:   "desabilita cores na saída",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Value:   20,
			},
		},
		HelpName: "go_search",
		Action: func(c *cli.Context) error {
			nargs := c.NArg()
			if nargs < 1 {
				_ = cli.ShowAppHelp(c)
				return nil
			}

			limit := c.Int("limit")
			plain = c.Bool("plain")
			query := strings.Join(c.Args().Slice(), " ")
			root := c.String("directory")
			start := time.Now()

			inc := c.StringSlice("include")

			filter := search.Filter{
				Root:     root,
				Include:  inc,
				SkipHide: !c.Bool("all"),
			}
			ch := filter.Run()

			srch := search.New{
				Filter: ch,
				Query:  query,
			}

			out := srch.Run()

			for r := range out {
				printResult(limit, r)
			}

			fmt.Printf("%s\n", time.Since(start))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
func max(j int, k int) int {
	if j > k {
		return j
	}
	return k
}

func min(j int, k int) int {
	if j < k {
		return j
	}
	return k
}

func printResult(limit int, r search.Result) {
	fmt.Println(clPathResult(r.Path) + ": ")

	var printStr string
	s := r.Line
	if limit == -1 {
		printStr = s
	} else {
		fstMatch := strings.Index(s, r.Query)
		lstMatch := strings.LastIndex(s, r.Query)

		left := max(fstMatch-limit, 0)
		right := min(lstMatch+len(r.Query)+limit, len(s))
		printStr = s[left:right]
	}
	printStr = strings.ReplaceAll(printStr, r.Query, clResult(r.Query))
	fmt.Println("  ", printStr)

}
