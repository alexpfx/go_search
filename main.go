package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"index/suffixarray"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "inclui na busca diretórios invisíveis",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "include pattern: se especificado irá incluir na busca apenas arquivos que satisfaçam o pattern",
			},
		},
		HelpName: "go_search",
		Action: func(c *cli.Context) error {
			nargs := c.NArg()
			if nargs < 1 {
				_ = cli.ShowAppHelp(c)
				return nil
			}
			query := strings.Join(c.Args().Slice(), " ")
			root := c.String("directory")

			start := time.Now()

			var incRegex *regexp.Regexp
			inc := c.String("include")
			if inc != "" {
				incRegex = regexp.MustCompile(inc)
			}

			ch := filter(root, c.Bool("all"), incRegex)
			out := search(ch, query)

			for r := range out {
				fmt.Println(r)
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

func search(in <-chan summary, query string) <-chan string {
	out := make(chan string, 4)

	go func() {
		for summary := range in {
			if summary.size < 1024*4 {
				small := searchSmall(summary, []byte(query))

				if small != ""{
					out <- small
				}
			}

			/*
				if bytes.Index(bs, []byte(query)) == -1 {
					continue
				}
			*/

			/*
				file, err := os.Open(summary)
				if err != nil {
					fmt.Println(err)
					continue
				}

				reader := bufio.NewReader(file)

				for {


					b, _, err := reader.ReadLine()
					//	fmt.Println(text)
					text := string(b)

					if err != nil {
						break
					}


					if strings.Contains(text, query) {
						out <- summary + "\n" + "->" + text
					}
					if err == io.EOF {
						break
					}

				}
				_ = file.Close()
			*/
		}

		close(out)
	}()

	return out
}

func searchSmall(summary summary, query []byte) string {
	bs, err := ioutil.ReadFile(summary.path)
	qLen := len(query)
	if err != nil {
		return ""
	}

	index := suffixarray.New(bs)

	lkpRes := index.Lookup(query, -1)
	if lkpRes == nil {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString(summary.path + "\n")

	for _, ix := range lkpRes {
		i := ix
		j := ix
		for {
			foundLeft := bs[i] == '\n' || i == 0
			if !foundLeft {
				i--
			}
			foundRight := bs[j] == '\n' || j == qLen
			if !foundRight {
				j++
			}

			if foundLeft && foundRight {
				break
			}
		}
		sb.Write(bs[i+1 : j-1])
		sb.WriteString("\n")
	}
	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type summary struct {
	path string
	size int64
}

func filter(root string, all bool, incRegex *regexp.Regexp) chan summary {
	out := make(chan summary, 4)

	go func() {

		_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path == root {
				return nil
			}

			if shouldSkipDir(entry, all) {
				return fs.SkipDir
			}

			if entry.IsDir() {
				return nil
			}

			if !entry.Type().IsRegular() {
				return nil
			}

			if incRegex != nil && !incRegex.MatchString(entry.Name()) {
				return nil
			}

			info, err := entry.Info()
			if err != nil {
				fmt.Println("erro " + err.Error())
				return nil
			}

			out <- summary{
				path: path,
				size: info.Size(),
			}

			return nil
		})
		close(out)
	}()

	return out
}

func shouldSkipDir(info fs.DirEntry, all bool) bool {
	if all {
		return false
	}
	if !info.IsDir() {
		return false
	}

	return strings.HasPrefix(info.Name(), ".")
}
