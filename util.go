package slurp

import (
	"bytes"

	"github.com/omeid/slurp/tools/glob"
)

//A Filter stage, will either close or pass files to the next
// Stage based on the output of the `filter` function.
func FilterFunc(c *C, filter func(File) bool) Stage {
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

// A simple transformation Stage, sends the file to output
// channel after passing it through the the "do" function.
func DoFunc(c *C, do func(*C, File) File) Stage {
	return func(files <-chan File, out chan<- File) {
		for f := range files {
			out <- do(c, f)
		}
	}
}

//For The Glory of Debugging.
func List(c *C) Stage {
	return DoFunc(c, func(c *C, f File) File {
		s, err := f.Stat()
		if err != nil {
			c.Print("Can't get file name.")
		} else {
			c.Printf("File: %+v Name: %s", f, s.Name())
		}
		return f
	})
}

//Filters out files based on a pattern, if they match,
// they will be closed, otherwise sent to the output channel.
func Filter(c *C, pattern string) Stage {
	return FilterFunc(c, func(f File) bool {
		s, err := f.Stat()
		if err != nil {
			c.Print("Can't get file name.")
			return false
		}
		m, err := glob.Match(pattern, s.Name())
		if err != nil {
			c.Println(err)
		}
		return m
	})
}

// Concatenates all the files from the input channel
// and passes them to output channel with the given name.
func Concat(c *C, name string) Stage {
	return func(files <-chan File, out chan<- File) {

		var (
			size    int64
			bigfile = new(bytes.Buffer)
		)

		for f := range files {
			n, err := bigfile.ReadFrom(f)
			if err != nil {
				c.Println(err)
				return
			}
			bigfile.WriteRune('\n')
			size += n + 1

			f.Close()
		}

		out <- File{
			Reader: bigfile,
			Dir:    "",
			Path:   name,
			stat:   &FileInfo{size: size, name: name},
		}
	}
}
