package slurp

import "github.com/omeid/slurp/log"

type C struct {
	log.Log
}



// A stage where a series of files goes for transformation, manipulation.
// There is no correlation between a stages input and output, a stage may
// decided to pass the same files after transofrmation or generate new files
// based on the input.


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

// Pipes the current Channel to the give list of Stages and returns the 
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
