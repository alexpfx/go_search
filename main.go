package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	app := &cli.App{
		HideHelpCommand: true,
		Name:  "go_search",
		Usage: "buscador de texto em arquivos",
		ArgsUsage: "<query_string>",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "directory",
				Aliases: []string{"d"},
				Usage: "diretório raíz de onde a busca deve partir",
				Value: "./",
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

			fmt.Println("root: ", root)
			fmt.Println("query: ", query)
			start := time.Now()

			ch := filter(root)
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
	out := make(chan string)

	go func() {
		for path := range in {
			file, err := os.Open(path)
			if err != nil {
				log.Println("não pode abrir arquivo " + path)
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				text := scanner.Text()

				if !strings.Contains(text, query) {
					continue
				}
				out <- path + "\n" + "->" + text
			}
			_ = file.Close()
		}
		close(out)
	}()

	return out
}

/*
func walk(root string) chan string {

	var wg sync.WaitGroup

	_ = filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		if !info.Type().IsRegular() {
			return nil
		}
		wg.Add(1)

		return nil
	})

	return nil

}
 */

func filter(root string) chan string {
	out := make(chan string)
	go func() {
		_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			out <- path
			return nil
		})
		close(out)
	}()
	return out
}
