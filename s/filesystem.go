package s

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/omeid/slurp/s/glob"
)

type filters []string

func (f filters) match(s string) bool {
	for _, f := range f {
		if match, _ := path.Match(f, s); match {
			return true
		}
	}
	return false
}

func parseglobs(globs ...string) map[string]string {
	return nil
}

func Src(globs ...string) Pipe {

	pipe := make(chan File)

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
		for _, glob := range globs {

			for _, file := range files {

				// if filters.match(file) { continue }
				file = filepath.Join(glob, file)
				_ = cwd
				//log.Printf("file %+v\n", file)
				//f, err := os.Open(file)
				//if err != nil {
				//	log.Println(err)
				//	return
				// }
				pipe <- File{Base: file}
			}

		}
	}()

	return pipe
}

func Dest(dst string) Job {
	return func(files <-chan File, out chan<- File) {
		for f := range files {
			log.Println(filepath.Join(dst, f.Base))
			f.Close()
		}
	}
}

func List(files <-chan File, out chan<- File) {
	for f := range files {
		log.Println(filepath.Join(f.Base))
		out <- f
	}
}
