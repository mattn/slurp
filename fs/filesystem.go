package fs

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/omeid/slurp/glob"
	"github.com/omeid/slurp/s"
)

func Src(globs ...string) s.Pipe {

	pipe := make(chan s.File)

	files, err := glob.Glob(globs...)

	if err != nil {
		log.Println(err)
		close(pipe)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		close(pipe)
		return pipe
	}

	//TODO: Parse globs here, check for invalid globs, split them into "filters".

	go func() {
		defer close(pipe)

		for file, pattern := range files {

			f, err := os.Open(file)
			if err != nil {
				log.Println(err)
				return
			}
			Stat, err := f.Stat()
			if err != nil {
				log.Println(err)
				return
			}
			pipe <- s.File{Cwd: cwd, Dir: glob.Dir(pattern), Path: file, Stat: Stat, Content: f}
		}

	}()

	return pipe
}

func Dest(dst string) s.Job {
	return func(files <-chan s.File, out chan<- s.File) {
		for file := range files {

			path := filepath.Join(dst, file.Dir)
			err := os.MkdirAll(path, 0700)
			if err != nil {
				log.Println(err)
				return
			}

			realfile, err := os.Create(filepath.Join(path, file.Stat.Name()))
			if err != nil {
				log.Println(err)
				return
			}
			io.Copy(realfile, file.Content)
			realfile.Close()
			out <- file
		}
	}
}
