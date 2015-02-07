package slurp

import "github.com/omeid/slurp/log"

type C struct {
	log.Log
}

type Job func(<-chan File, chan<- File)

func (j Job) pipe(p <-chan File) Pipe {
	//func (p Pipe) Pipe(j Job) Pipe {
	out := make(chan File)
	go func() {
		defer close(out)
		j(p, out)
	}()

	return out
}

type Pipe <-chan File

func (p Pipe) Pipe(j ...Job) Pipe {
	switch len(j) {
	case 0:
		return p
	case 1:
		return j[0].pipe(p)
	default:
		return j[0].pipe(p).Pipe(j[1:]...)
	}
}

func (p Pipe) Wait() error {
	//Waits for the "build" to Pipe to finish and closes all
	// files, returns the first error.
	var err error
	for f := range p {
		e := f.Close()
		if err == nil && e != nil {
			err = e
		}
	}
	return err
}
