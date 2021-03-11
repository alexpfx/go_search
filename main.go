package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"io/fs"
	"log"
	"net/http"
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

func search(in <-chan string, query string) <-chan string {
	out := make(chan string, 4)

	go func() {
		for path := range in {
			file, err := os.Open(path)
			if err != nil {
				continue
			}

			buff := make([]byte, 512)
			rd := bufio.NewReader(file)
			_, err = rd.Read(buff)
			if err == io.EOF {
				_ = file.Close()
				fmt.Println(path)
				continue
			}
			contentType := http.DetectContentType(buff)
			if !strings.HasPrefix(contentType, "text") {
				_ = file.Close()
				continue
			}

			text := searchAll(file, []byte(query))

			if text != "" {
				out <-
					fmt.Sprintf("\n\n%s\n    %s\n", path, text)
			}
		}

		close(out)
	}()

	return out
}

func searchAll(file *os.File, query []byte) string {
	r := bufio.NewReader(file)

	scanner := bufio.NewScanner(r)
	sb := strings.Builder{}
	for scanner.Scan() {
		text := scanner.Text()

		if !strings.Contains(text, string(query)) {
			continue
		}

		sb.WriteString("\n  ")
		sb.WriteString(strings.TrimSpace(text))

	}
	_ = file.Close()
	return sb.String()
}

func filter(root string, all bool, incRegex *regexp.Regexp) chan string {
	out := make(chan string, 4)

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
				fmt.Println("not regular")
				return nil
			}

			out <- path

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
