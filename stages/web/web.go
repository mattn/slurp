package web

import (
	"mime"
	"net/http"
	"path"

	"github.com/omeid/slurp"
)

func name(url string, response *http.Response) string {

	_, params, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))

	name, ok := params["filename"]
	if !ok || err != nil {
		name = path.Base(url)
	}

	return name
}

func Get(c *slurp.C, urls ...string) slurp.Pipe {

	pipe := make(chan slurp.File)

	go func() {
		defer close(pipe)

		for _, url := range urls {

			c.Printf("Downloading %s", url)
			resp, err := http.Get(url)
			if err != nil {
				c.Println(err)
				break
			}

			if resp.StatusCode < 200 || resp.StatusCode > 399 {
				c.Printf("%s (%s)", resp.Status, url)
				continue
			}

			name := name(url, resp)

			content := c.ReadProgress(resp.Body, "Downloading "+name, resp.ContentLength)

			Stat := &slurp.FileInfo{}
			Stat.SetName(name)
			Stat.SetSize(resp.ContentLength)

			pipe <- slurp.File{Reader: content, Cwd: "", Dir: "", Path: name, Stat: Stat}
		}
	}()

	return pipe
}

/*
func Put(url url.URL) slurp.Job {
	return func(files <-chan slurp.File, out chan<- slurp.File) {
		for file := range files {
			_ = file
			/*
			// */ /*
		}
	}
}

*/
