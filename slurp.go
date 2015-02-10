package slurp

import "github.com/omeid/slurp/log"

type C struct {
	log.Log
}


type Stage func(<-chan File, chan<- File)

func (j Stage) pipe(p <-chan File) Pipe {
	//func (p Pipe) Pipe(j Stage) Pipe {
	out := make(chan File)
	go func() {
		defer close(out)
		j(p, out)
	}()

	return out
}


//Pipe is a channel of Files.
type Pipe <-chan File

// Pipes the current channel to the give list of jobs and returns the 
// last jobs otput pipe.
func (p Pipe) Pipe(j ...Stage) Pipe {
	switch len(j) {
	case 0:
		return p
	case 1:
		return j[0].pipe(p)
	default:
		return j[0].pipe(p).Pipe(j[1:]...)
	}
}

// Waits for the end of channel and closes all the files.
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

//This is a combination of p.Pipe(....).Wait()
func (p Pipe) Then(j ...Stage) error {
  return p.Pipe(j...).Wait()
}

