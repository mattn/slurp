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

type MatchPair struct {
	Glob string
	Name string
}

type pattern struct {
	Glob     string
	Negative bool
}

func excluded(i int, p []pattern, name string) bool {

	for i := i; i < len(p); i++ {
		if !p[i].Negative {
			continue
		}
		if m, _ := filepath.Match(p[i].Glob, name); m {
			return true
		}
	}
	return false
}

func Glob(globs ...string) (<-chan MatchPair, error) {

	//defer close(out)

	patterns := []pattern{}

	for _, glob := range globs {
		negative, err := Match(glob, "")
		if err != nil {
			return nil, err
		}

		if negative {
			glob = glob[1:]
		}

		patterns = append(patterns, pattern{glob, negative})
	}

	matches := make(chan MatchPair)
	go func() {

		seen := make(map[string]struct{})

		defer close(matches)
		for i, pattern := range patterns {

			if pattern.Negative {
				continue
			}
			//Patterns already checked and fs errors are ignored. so no error handling here.
			files, _ := filepath.Glob(pattern.Glob)

			for _, file := range files {
				if _, seen := seen[file]; seen || excluded(i, patterns, file) {
					continue
				}

				seen[file] = struct{}{}
				matches <- MatchPair{pattern.Glob, file}
			}
		}
	}()

	return matches, nil
}
