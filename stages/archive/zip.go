// Package archive provides archiving Stages for Slurp.

package archive

import (
	"bytes"
	"io/ioutil"
	"sync"

	"archive/zip"

	"github.com/omeid/slurp"
)

// Unzip the zip files from input channel and pass the result
// to the output channel. 
func Unzip(c *slurp.C) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		var wg sync.WaitGroup
		for file := range in {

			wg.Add(1)
			go func(file slurp.File) {
				defer wg.Done()

				raw, err := ioutil.ReadAll(file)
				file.Close()

				r, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
				if err != nil {
					c.Println(err)
					return
				}

				counter := c.Counter("unzipping", len(r.File))

				// Iterate through the files in the archive,
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
