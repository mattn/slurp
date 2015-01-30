package fs

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/omeid/slurp/glob"
	"github.com/omeid/slurp/s"
)

func Src(c *s.C, globs ...string) s.Pipe {

	pipe := make(chan s.File)

	files, err := glob.Glob(globs...)

	if err != nil {
		c.Println(err)
		close(pipe)
	}

	cwd, err := os.Getwd()
	if err != nil {
		c.Println(err)
		close(pipe)
		return pipe
	}

	//TODO: Parse globs here, check for invalid globs, split them into "filters".
	go func() {
		defer close(pipe)

		for matchpair := range files {
			f, err := os.Open(matchpair.Name)
			if err != nil {
				c.Println(err)
				return
			}
			Stat, err := f.Stat()
			if err != nil {
				c.Println(err)
				return
			}
			pipe <- s.File{Reader: f, Cwd: cwd, Dir: glob.Dir(matchpair.Glob), Path: matchpair.Name, Stat: Stat}
		}

	}()

	return pipe
}

func Dest(c *s.C, dst string) s.Job {
	return func(files <-chan s.File, out chan<- s.File) {

		var wg sync.WaitGroup
		defer wg.Wait()

		for file := range files {

			realpath, _ := filepath.Rel(file.Dir, file.Path)
			path := filepath.Join(dst, filepath.Dir(realpath))
			err := os.MkdirAll(path, 0700)
			if err != nil {
				//c.Println(err)
				return
			}

			if !file.Stat.IsDir() {

				realfile, err := os.Create(filepath.Join(dst, realpath))

				if err != nil {
					c.Println(err)
					return
				}

				wg.Add(1)
				go func(realfile *os.File, file io.Reader) {
					defer realfile.Close()
					defer wg.Done()

					io.Copy(realfile, file)
					realfile.Close()
				}(realfile, file)
			}
			out <- file
		}

	}
}
