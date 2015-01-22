// +build ignore
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/omeid/slurp/s"
)

/*
func init() {
	fmt.Println("_slurp: init.")
}
*/

func Slurp() error {

	//TODO: slurp needs to check if this function depends on the main package and import it.
	//	_ = dep
	fmt.Println("_s.Slurp")

	b := s.NewBuild()

	b.Task("say-hello", nil, func() error {
		wait := time.Duration(rand.Intn(10)) * time.Second
		log.Printf("HELLO. (wait %s.)", wait)
		time.Sleep(wait)
		log.Printf("BYE.   (wait %s.)", wait)
		return nil
	})

	b.Task("testing", []string{"say-hello"}, func() error {
		s.Src([]string{"./**"}).Pipe(
			func(files <-chan s.File, out chan<- s.File) {
				for f := range files {
					fmt.Printf("--> %s\n", f.Base)
					out <- f
				}
			}).Pipe(s.Dist("/public")).Wait()

		return nil
	})

	b.Run("say-hello", "testing", "say-hello", "say-hello", "say-hello").Wait()

	return nil
}
