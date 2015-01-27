package template

import (
	"bytes"
	html "html/template"
	"io"
	"sync"

	"github.com/omeid/slurp/s"
)

type executable interface {
	Execute(io.Writer, interface{}) error
}

func NewTemplateReadCloser(c *s.C, wg sync.WaitGroup, e executable, data interface{}) templateReadCloser {

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

func HTML(c *s.C, data interface{}) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		templates := html.New("")

		var wg sync.WaitGroup
		defer wg.Wait() //Wait before all templates are executed.

		for f := range in {

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(f.Content)
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

			f.Content = NewTemplateReadCloser(c, wg, template, data)

			out <- f
		}
	}
}
