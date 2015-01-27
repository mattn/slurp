package s

import (
	"bytes"

	"github.com/omeid/slurp/glob"
)

func FilterFunc(c *C, filter func(File) bool) Job {
	return func(files <-chan File, out chan<- File) {
		for f := range files {
			if filter(f) {
				f.Close()
			} else {
				out <- f
			}
		}
	}
}

func DoFunc(c *C, do func(*C, File) File) Job {
	return func(files <-chan File, out chan<- File) {
		for f := range files {
			out <- do(c, f)
		}
	}
}

//For The Glory of Debugging.
func List(c *C) Job {
	return DoFunc(c, func(c *C, f File) File {
		c.Printf("File: %+v ", f)
		return f
	})
}

func Filter(c *C, pattern string) Job {
	return FilterFunc(c, func(f File) bool {
		m, err := glob.Match(pattern, f.Stat.Name())
		if err != nil {
			c.Println(err)
		}
		return m
	})
}

func Concat(c *C, output string) Job {
	return func(files <-chan File, out chan<- File) {

		var (
			size    int64
			bigfile = new(bytes.Buffer)
		)

		for f := range files {
			n, err := bigfile.ReadFrom(f.Content)
			if err != nil {
				c.Println(err)
				return
			}
			size += n

			f.Close()
		}

		out <- File{
			Dir:     "",
			Path:    output,
			Stat:    &FileInfo{size: size, name: output},
			Content: bigfile,
		}
	}
}
