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
	if pattern != "" && pattern[0] == '!' {
		match, err := filepath.Match(pattern[1:], name)
		return !match, err
	}

	return filepath.Match(pattern, name)
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

func Glob(globs ...string) ([]string, error) {

	matches := []string{}

	for _, glob := range globs {

		isNot, err := Match(glob, "")
		if err != nil {
			return nil, err
		}

		if isNot {
			matches, _ = Filter(glob, matches)
		} else {

			files, err := filepath.Glob(glob)
			if err != nil {
				return matches, err
			}

			matches = append(matches, files...)
		}
	}

	return matches, nil
}
