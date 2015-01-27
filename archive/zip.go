package archive

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"sync"

	"archive/zip"

	"github.com/omeid/slurp/s"
)

func Unzip(c *s.C) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		//Because zip is not an streaming archive, we don't want to block.
		var wg sync.WaitGroup
		for file := range in {

			go func(file s.File) {
				wg.Add(1)
				defer wg.Done()

				raw, err := ioutil.ReadAll(file.Content)
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

					name := filepath.Base(f.Name)
					//content = multibar.NewReadCloserHook(content, bars.MakeBar(int(f.FileInfo().Size()), name))

					out <- s.File{Dir: filepath.Dir(f.Name), Path: name, Stat: f.FileInfo(), Content: content}

				}
			}(file)

		}
		wg.Wait()
	}
}
