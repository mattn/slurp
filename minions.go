package m

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

type Pipe chan File

type Job func(<-chan File, <-chan File)

func (p *Pipe) Pipe(j Job) Pipe {
	out := make(Pipe)
	go func() {
		defer close(out)
		j((<-chan File)(*p), out)
	}()
	return out
}

func Src(globs []string) *Pipe {

	pipe := make(Pipe)

	go func() {
		defer close(pipe)
		for _, glob := range globs {
			files, err := filepath.Glob(glob)
			if err != nil {
				log.Println(err)
				return
			}
			for _, file := range files {
				fmt.Printf("file %+v\n", file)
				pipe <- File{Base: file}
			}

		}
	}()

	return &pipe
}

func Dist(dst string) Job {
	return func(files <-chan File, out <-chan File) {
		for f := range files {
			fmt.Println(filepath.Join(dst, f.Base))
		}
	}
}
