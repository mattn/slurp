package archive

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"archive/zip"

	"github.com/omeid/slurp/s"
)

func Unzip(c *s.C) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		for file := range in {

			raw, err := ioutil.ReadAll(file.Content)
			file.Close()

			r, err := zip.NewReader(bytes.NewReader(raw), file.Stat.Size())
			if err != nil {
				c.Println(err)
				continue
			}

			// Iterate through the files in the archive,
			// printing some of their contents.
			counter := c.Counter("unzipping", len(r.File))
			//bars := c.New("    ")
			for i, f := range r.File {
				counter.Set(i, f.Name)

				content, err := f.Open()
				if err != nil {
				}

				name := filepath.Base(f.Name)
				//content = multibar.NewReadCloserHook(content, bars.MakeBar(int(f.FileInfo().Size()), name))

				out <- s.File{Dir: filepath.Dir(f.Name), Path: name, Stat: f.FileInfo(), Content: content}

			}
		}
	}
}
