package watch

import (
	"log"
	"sync"

	"github.com/omeid/slurp/glob"
	"github.com/omeid/slurp/s"
	"gopkg.in/fsnotify.v1"
)

func Watch(b *s.Build, c *s.C, wg *sync.WaitGroup, task string, globs ...string) {

	wg.Add(1)
	go func() {
		defer wg.Done()

		files, err := glob.Glob(globs...)

		if err != nil {
			c.Println(err)
			return
		}

		w, err := fsnotify.NewWatcher()
		if err != nil {
			c.Println(err)
			return
		}

		for matchpair := range files {
			w.Add(matchpair.Name)
		}

		b.Defer(func() {
			c.Printf("Stopping watch for %s.", task)
			w.Close()
		})

		for {
			select {
			case event := <-w.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					b.Run([]string{task})
				}
			case err := <-w.Errors:
				c.Println(err)
			}
		}
	}()

}
