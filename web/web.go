package web

import (
	"mime"
	"net/http"
	"net/url"
	"path"

	"github.com/omeid/slurp/s"
)

func name(url string, response *http.Response) string {

	_, params, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))

	name, ok := params["filename"]
	if !ok || err != nil {
		name = path.Base(url)
	}

	return name
}

func Get(c *s.C, urls ...string) s.Pipe {

	pipe := make(chan s.File)

	go func() {
		defer close(pipe)

		for _, url := range urls {

			c.Printf("Downloading %s", url)
			resp, err := http.Get(url)
			if err != nil {
				c.Println(err)
				break
			}
			name := name(url, resp)

			content := c.ReadProgress(resp.Body, "Downloading "+name, resp.ContentLength)

			Stat := &s.FileInfo{}
			Stat.SetName(name)
			Stat.SetSize(resp.ContentLength)

			pipe <- s.File{Cwd: "", Dir: "", Path: name, Stat: Stat, Content: content}
		}
	}()

	return pipe
}

func Put(url url.URL) s.Job {
	return func(files <-chan s.File, out chan<- s.File) {
		for file := range files {
			_ = file
			/*
				path := filepath.Join(dst, file.Dir)
				err := os.MkdirAll(path, 0700)
				if err != nil {
					log.Println(err)
					return
				}

				realfile, err := os.Create(filepath.Join(path, file.Stat.Name()))
				if err != nil {
					log.Println(err)
					return
				}
				io.Copy(realfile, file.Content)
				realfile.Close()
				out <- file
			*/
		}
	}
}
