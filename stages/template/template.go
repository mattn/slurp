//package template provides Go Templates compiling Stages for Slurp.

package template

import (
	"bytes"
	html "html/template"
	"io"
	"sync"

	"github.com/omeid/slurp"
)

type executable interface {
	Execute(io.Writer, interface{}) error
}

func NewTemplateReadCloser(c *slurp.C, wg sync.WaitGroup, e executable, data interface{}) templateReadCloser {

	buf := new(bytes.Buffer)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := e.Execute(buf, data)
		if err != nil {
			c.Println(err)
		}
	}()

	return templateReadCloser{buf}
}

type templateReadCloser struct {
	io.Reader
}

func (t templateReadCloser) Close() error {
	return nil
}


// Compiles the input files a html/template using the provided data
// and passes them down the line.
// It creates an increamental collection of templates that allows accessing
// templates from templates (Just pass the "required" templates first.)
func HTML(c *slurp.C, data interface{}) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		templates := html.New("")

		var wg sync.WaitGroup
		defer wg.Wait() //Wait before all templates are executed.

		for f := range in {

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(f.Reader)
			f.Close()
			if err != nil {
				c.Println(err)
				break
			}

			template, err := templates.New(f.Stat.Name()).Parse(buf.String())
			if err != nil {
				c.Println(err)
				break
			}

			f.Reader = NewTemplateReadCloser(c, wg, template, data)

			out <- f
		}
	}
}
