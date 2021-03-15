package main

import (
	"fmt"
	"github.com/alexpfx/go_search/search"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var clCyan = color.New(color.FgHiCyan)
var clBlue = color.New(color.FgHiBlue)
var plain bool

func blue(a interface{}) {
	clOut(clBlue, plain, a)
}

func cyan(a interface{}) {
	clOut(clCyan, plain, a)
}

func clOut(color *color.Color, plain bool, a interface{}) {
	if plain {
		fmt.Print(a)
		return
	}
	_, _ = color.Print(a)
}

func main() {
	app := &cli.App{
		HideHelpCommand: true,
		Name:            "go_search",
		Usage:           "buscador de texto em arquivos",
		ArgsUsage:       "<query_string>",

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
			&cli.StringFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "include pattern: se especificado irá incluir na busca apenas arquivos que satisfaçam o pattern",
			},
			&cli.BoolFlag{
				Name:    "plain",
				Aliases: []string{"p"},
				Value:   false,
				Usage:   "desabilita cores na saída",
			},
		},
		HelpName: "go_search",
		Action: func(c *cli.Context) error {
			nargs := c.NArg()
			if nargs < 1 {
				_ = cli.ShowAppHelp(c)
				return nil
			}

			plain = c.Bool("plain")
			query := strings.Join(c.Args().Slice(), " ")
			root := c.String("directory")

			start := time.Now()

			var incRegex *regexp.Regexp
			inc := c.String("include")
			if inc != "" {
				incRegex = regexp.MustCompile(inc)
			}

			filter := search.Filter{
				Root:     root,
				Include:  incRegex,
				SkipHide: !c.Bool("all"),
			}
			ch := filter.Run()

			srch := search.New{
				Filter: ch,
				Query:  query,
			}

			out := srch.Run()

			for r := range out{
				printResult(r)
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

func printResult(r search.Result) {
	cyan(r.Path + ": ")

	for i, s := range strings.Split(r.Line, r.Query) {
		if i != 0 {
			blue(r.Query)
		}
		fmt.Print(s)
	}
	fmt.Println()
}
