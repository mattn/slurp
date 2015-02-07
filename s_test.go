package slurp

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestSrc(t *testing.T) {
	Src([]string{"../**", "../exmample/**/*.go"}).Wait()
}

func TestxSrci(t *testing.T) {

	b := NewBuild()

	b.Task("say-hello", nil, func() error {
		wait := time.Duration(rand.Intn(10)) * time.Second
		log.Printf("HELLO. I am going to take %s.", wait)
		time.Sleep(wait)
		return nil
	})

	b.Task("testing", []string{"say-hello"}, func() error {
		Src([]string{"./**"}).Pipe(
			func(files <-chan File, out chan<- File) {
				for f := range files {
					fmt.Printf("--> %s\n", f.Base)
					out <- f
				}
			}).Pipe(Dist("/public")).Wait()

		return nil
	})

	b.Task("default", []string{"say-hello", "testing", "say-hello", "say-hello", "say-hello"}, func() error {
		return nil
	})

	b.Run("default").Wait()
}
