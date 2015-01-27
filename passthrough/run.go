package passthrough

import (
	"os"
	"os/exec"
	"sync"

	"github.com/omeid/slurp/s"
)

func Run(c *s.C, name string, args ...string) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		//Because programs block, zip is not an streaming archive, we don't want to block.
		var wg sync.WaitGroup
		defer wg.Wait()

		for file := range in {

			cmd := exec.Command(name, args...)
			cmd.Stderr = os.Stderr //TODO: io.Writer logger.

			cmd.Stdin = file.Content
			content, err := cmd.StdoutPipe()
			if err != nil {
				c.Println(err)
				return
			}

			err = cmd.Start()
			if err != nil {
				c.Println(err)
				return
			}

			wg.Add(1)
			go func(cmd *exec.Cmd) {
				defer wg.Done()
				defer s.Close(content)
				cmd.Wait()
			}(cmd)

			file.Content = content
			out <- file

		}

	}
}
