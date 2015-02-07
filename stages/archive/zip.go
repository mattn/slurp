package archive

import (
	"bytes"
	"io/ioutil"
	"sync"

	"archive/zip"

	"github.com/omeid/slurp"
)

func Unzip(c *slurp.C) slurp.Job {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		var wg sync.WaitGroup
		for file := range in {

			go func(file slurp.File) {
				wg.Add(1)
				defer wg.Done()

				raw, err := ioutil.ReadAll(file)
				file.Close()

				r, err := zip.NewReader(bytes.NewReader(raw), file.Stat.Size())
				if err != nil {
					c.Println(err)
					return
				}

				// Iterate through the files in the archive,
				// printing some of their contents.
				counter := c.Counter("unzipping", len(r.File))
				//bars := c.New("    ")
				for i, f := range r.File {
					counter.Set(i+1, f.Name)

					content, err := f.Open()
					if err != nil {
					}
					out <- slurp.File{Reader: content, Dir: "", Path: f.Name, Stat: f.FileInfo()}

				}
			}(file)

		}
		wg.Wait()
	}
}
