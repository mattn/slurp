package s

import (
	"io"
	"os"
)

type File struct {
	Cwd     string
	Base    string
	Path    string
	Stat    os.FileInfo
	Content io.Reader
}

func (f *File) Close() error {
	if closer, ok := f.Content.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type Job func(<-chan File, chan<- File)

//func (j Job) run(p <-chan File) Pipe {
func (p Pipe) Pipe(j Job) Pipe {
	out := make(chan File)
	go func() {
		defer close(out)
		j(p, out)
	}()

	return out
}

type Pipe <-chan File

/*
func (p Pipe) Pipe(j ...Job) Pipe {
	switch len(j) {
	case 0:
		return p
	case 1:
		return j[0].run(p)
	default:
		return j[0].run(p).Pipe(j[1:]...)
	}
}
*/

func (p Pipe) Wait() {
	for f := range p {
		f.Close()
	}
}
