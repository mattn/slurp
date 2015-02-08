package exec

// Package exec runs external commands. It wraps os/exec to make it
// easier to copy and kill a os/exec.Cmd

import (
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Cmd struct {
	exec.Cmd

	bin    string
	params []string
}

func Command(bin string, args ...string) *Cmd {
	return &Cmd{
		Cmd:    *exec.Command(bin, args...),
		bin:    bin,
		params: args,
	}
}

func (r *Cmd) New() *Cmd {
	cmd := Command(r.bin, r.params...)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr

	return cmd

}

func (r *Cmd) Kill() error {
	if r.Cmd.Process != nil {
		done := make(chan error)
		go func() {
			r.Cmd.Wait()
			close(done)
		}()
		//Trying a "soft" kill first
		var err error
		if runtime.GOOS == "windows" {
			err = r.Cmd.Process.Kill()
		} else {
			err = r.Cmd.Process.Signal(os.Interrupt)
		}
		if err != nil {
			return err
		}
		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			return r.Cmd.Process.Kill()
		case <-done:
		}
	}
	return nil
}
