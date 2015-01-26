package template

import (
	"bytes"
	html "html/template"
	"io"
	"log"
	"sync"

	"github.com/omeid/slurp/s"
)

type executable interface {
	Execute(io.Writer, interface{}) error
}

func NewTemplateReadCloser(wg sync.WaitGroup, e executable, data interface{}) templateReadCloser {

	buf := new(bytes.Buffer)
	go func() {
		wg.Add(1)
		defer wg.Done()
		e.Execute(buf, data)
	}()

	return templateReadCloser{buf}
}

type templateReadCloser struct {
	io.Reader
}

func (t templateReadCloser) Close() error {
	return nil
}

//interface { New() New() Parse() Execute()
func HTML(data interface{}) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		templates := html.New("")

		var wg sync.WaitGroup
		defer wg.Wait() //Wait before all templates are executed.

		for f := range in {

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(f.Content)
			f.Close()
			if err != nil {
				log.Println(err)
				break
			}

			template, err := templates.New(f.Stat.Name()).Parse(buf.String())
			if err != nil {
				log.Println(err)
				break
			}

			f.Content = NewTemplateReadCloser(wg, template, data)

			out <- f
		}
	}
}
