package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	app := cli.App{
		Action: func(c *cli.Context) error {
			nargs := c.NArg()
			if nargs < 1 {
				_ = cli.ShowAppHelp(c)
				return nil
			}
			query := c.Args().Get(nargs - 1)
			root := "./"

			if nargs > 1 {
				root = c.Args().Get(0)
			}

			ctx, cf := context.WithTimeout(context.Background(), time.Second*120)
			defer cf()

			start := time.Now()
			_, err := search(ctx, root, query)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", time.Since(start))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func search(context context.Context, root string, query string) ([]string, error) {


	group, cx := errgroup.WithContext(context)

	chanPath := make(chan string)

	group.Go(func() error {
		defer close(chanPath)
		return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			select {
			case chanPath <- path:
			case <-cx.Done():
				return cx.Err()
			}

			return nil
		})
	})

	chanSearch := make(chan string, 4)
	for path := range chanPath {
		p := path
		group.Go(func() error {
			file, err := os.Open(p)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				text := scanner.Text()
				if !strings.Contains(text, query) {
					continue
				}

				select {
				case chanSearch <- fmt.Sprintf(" %s --> %s", text, p):
				case <-cx.Done():
					return cx.Err()
				}
			}

			return scanner.Err()
		})

	}

	go func() {
		group.Wait()
		close(chanSearch)
	}()

	var res []string
	for r := range chanSearch {
		fmt.Println(r)
		res = append(res, r)
	}
	return res, group.Wait()

}
