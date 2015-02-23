//package template provides Go Templates compiling Stages for Slurp.

package template

import (
	"bytes"
	"html/template"

	"github.com/omeid/slurp"
)

// Compiles the input files a html/template using the provided data
// and passes them down the line.
// It creates an increamental collection of templates that allows accessing
// templates from templates (Just pass the "required" templates first.)
func HTML(c *slurp.C, data interface{}) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		templates := template.New("")

		for f := range in {

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(f.Reader)
			f.Close()
			if err != nil {
				c.Println(err)
				break
			}

			s, err := f.Stat()
			if err != nil {
				c.Println("Can't get file name.")
				break
			}

			template, err := templates.New(s.Name()).Parse(buf.String())
			if err != nil {
				c.Println(err)
				break
			}

			buff := new(bytes.Buffer)
			err = template.Execute(buff, data)
			if err != nil {
			  c.Println(err)
			  break
			}

			f.Reader = buff

			out <- f
		}
	}
}
