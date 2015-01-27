package glob

import (
	"path/filepath"
	"strings"
)

func Dir(glob string) string {
	if strings.IndexAny(glob, "*?[") < 0 {
		return glob
	}
	return Dir(filepath.Dir(glob))
}

func Base(glob string) string {
	base, _ := filepath.Rel(Dir(glob), glob)
	return base
}

func Match(pattern, name string) (bool, error) {

  negative := pattern != "" && pattern[0] == '!'
  if negative {
	pattern = pattern[1:]
  }

  m, err := filepath.Match(pattern, name)
  return m != negative, err
}

func Filter(pattern string, files []string) ([]string, error) {
	isNot, err := Match(pattern, "")
	if err != nil {
		return nil, err
	}

	if isNot {
		pattern = pattern[1:]
	}

	out := []string{}

	for _, file := range files {
		if match, _ := filepath.Match(pattern, file); match != isNot {
			out = append(out, file)
		}
	}
	return out, nil
}

func Glob(globs ...string) (map[string]string, error) {

	matches := make(map[string]string)

	for _, glob := range globs {

		isNot, err := Match(glob, "")
		if err != nil {
			return nil, err
		}

		if isNot {
			glob = glob[1:]

			for file, _ := range matches {
				if match, _ := filepath.Match(glob, file); match {
					delete(matches, file)
				}
			}
		} else {

			files, err := filepath.Glob(glob)
			if err != nil {
				return matches, err
			}

			for _, file := range files {
				matches[file] = glob
			}
		}
	}

	return matches, nil
}
