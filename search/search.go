package search

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type New struct {
	Filter    chan string
	Query     string
	PlainText bool
}

func (s New) Run() <-chan Result {
	return search(s.Filter, s.Query)
}

func search(in <-chan string, query string) <-chan Result {
	out := make(chan Result, 4)

	go func() {
		for path := range in {

			file, err := os.Open(path)
			if err != nil {
				continue
			}

			tokens := scanAndMatch(file, query)

			err = file.Close()

			if len(tokens) == 0 {
				continue
			}

			for _, t := range tokens {
				out <- Result{
					Query: query,
					Line:  t,
					Path:  path,
				}
			}

		}
		close(out)
	}()

	return out
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

func scanAndMatch(file *os.File, query string) []string {
	r := bufio.NewReader(file)

	scanner := bufio.NewScanner(r)
	lines := make([]string, 0)


	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !strings.Contains(line, query) {
			continue
		}
		lines = append(lines, line)
	}

	return lines
}

type Filter struct {
	Root     string
	Include  *regexp.Regexp
	SkipHide bool
}

func (f Filter) Run() chan string {
	return filter(f.Root, !f.SkipHide, f.Include)
}

func filter(root string, all bool, incRegex *regexp.Regexp) chan string {
	out := make(chan string, 8)

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

			if incRegex != nil && !incRegex.MatchString(path){
				return nil
			}

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

type Result struct {
	Query string
	Line  string
	Path  string
}
