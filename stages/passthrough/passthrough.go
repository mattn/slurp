/*
passthrough allows you to pass files through an executable program.
The program is executed for every single file, If you want to pass a series 
of files through a single invocation of the program, please use slurp.Concat
and pipe it to passtrhough to be processed by your designed program.
*/
package passthrough

import (
	"os"
	"os/exec"
	"sync"

	"github.com/omeid/slurp"
)

// bin is the binary name, it will be passed to os/exec.Command, so the same
// path rules applies.
// the args are the argumetns passed to the program.
func Run(c *slurp.C, bin string, args ...string) slurp.Job {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		//Because programs block, zip is not an streaming archive, we don't want to block.
		var wg sync.WaitGroup
		defer wg.Wait()

		for file := range in {

			cmd := exec.Command(bin, args...)
			cmd.Stderr = os.Stderr //TODO: io.Writer logger.

			cmd.Stdin = file.Reader
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
				defer slurp.Close(content)
				cmd.Wait()
			}(cmd)

			file.Reader = content
			out <- file

		}

	}
}
