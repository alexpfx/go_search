package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	Black   = "\033[1;30m%s\033[0m"
	Red     = "\033[1;31m%s\033[0m"
	Green   = "\033[1;32m%s\033[0m"
	Yellow  = "\033[1;33m%s\033[0m"
	Purple  = "\033[1;34m%s\033[0m"
	Magenta = "\033[1;35m%s\033[0m"
	Teal    = "\033[1;36m%s\033[0m"
	White   = "\033[1;37m%s\033[0m"
)

var (
	PathColor      = Red
	HighlightColor = Yellow
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
			fmt.Println("buscando: ", path)
			file, err := os.Open(path)
			if err != nil {
				continue
			}

			tokens := searchAll(file, query)

			fmt.Println("fechando: ", path)
			err = file.Close()

			if len(tokens) == 0 {
				continue
			}


			out <- "encontrou " + path
			/*
				out <- fmt.Sprintf(Yellow, path) + "\n"
				for _, t := range tokens {
					out <- fmt.Sprintf("%s: %s",
						fmt.Sprintf(PathColor, path),
						strings.Replace(t, query,
							fmt.Sprintf(HighlightColor, query), 4))
				}
			*/

		}
		close(out)
	}()

	return out
}

func searchAll(file *os.File, query string) []string {
	r := bufio.NewReader(file)

	scanner := bufio.NewScanner(r)
	tokens := make([]string, 0)

	for scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())

		if !strings.Contains(token, query) {
			continue
		}
		tokens = append(tokens, token)
	}

	return tokens
}

func printSize(size int64) string {
	r := float64(size) / 1024
	if r < 1024 {
		return fmt.Sprintf("%2.2fKB", r)
	}
	r /= 1024
	if r < 1024 {
		return fmt.Sprintf("%2.2fMB", r)
	}

	r /= 1024
	return fmt.Sprintf("%2.2fGB", r)

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

			if shouldIgnoreExt(filepath.Ext(path)) {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return nil
			}


			fmt.Printf("enviando: %s %s\n", path, printSize(info.Size()))
			out <- path

			return nil
		})
		close(out)
	}()

	return out
}

func shouldIgnoreExt(ext string) bool {
	ignore := []string{"jar", "tar", "zip", "bin"}

	for _, iext := range ignore {
		if strings.EqualFold(iext, ext) {
			return true
		}
	}
	return false
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
