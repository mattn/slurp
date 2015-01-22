package s

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type File struct {
	Cwd     string
	Base    string
	Path    string
	Stat    os.FileInfo
	Content io.ReadWriteCloser
}

type Pipe <-chan File

type Job func(<-chan File, chan<- File)

func (p Pipe) Pipe(j Job) Pipe {
	out := make(chan File)
	go func() {
		defer close(out)
		j(p, out)
	}()
	return out
}

func (p Pipe) Wait() {
	for _ = range p {
	}
}

func Src(globs []string) Pipe {

	pipe := make(chan File)

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		close(pipe)
		return pipe
	}

	go func() {
		defer close(pipe)
		for _, glob := range globs {

			files, err := filepath.Glob(glob)
			if err != nil {
				log.Println(err)
				return
			}

			for _, file := range files {
				//f, err := os.Open(file)
				//if err != nil {
				//	log.Println(err)
				//		return
				//	}
				_ = cwd
				pipe <- File{Base: file}
			}

		}
	}()

	return pipe
}

func Dist(dst string) Job {
	return func(files <-chan File, out chan<- File) {
		for f := range files {
			fmt.Println(filepath.Join(dst, f.Base))
		}
	}
}
