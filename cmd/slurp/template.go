package main

import "text/template"

var runnerSrc = template.Must(template.New("main").Parse(`
package main

import (
  "flag"
  "strings"
  "os"
  "os/signal"
  "runtime"

  "github.com/omeid/slurp"

  client "{{ . }}/tmp"
)

func init() {

  maxprocs := runtime.NumCPU()
  if maxprocs > 2 {
  runtime.GOMAXPROCS(maxprocs/2)
  }
}

func main() {

  flag.Parse()
  slurp := slurp.NewBuild()

  interrupts := make(chan os.Signal, 1)
  signal.Notify(interrupts, os.Interrupt)

  go func() {
	sig := <-interrupts
	// stop watches and clean up.
	slurp.Printf("captured %v, stopping build and exiting..\n", sig)
	slurp.Close() 
	os.Exit(1)
  }()


  client.Slurp(slurp)

  tasks := flag.Args()
  if len(tasks) == 0 {
	tasks = []string{"default"}
  }

  slurp.Printf("Running: %s", strings.Join(tasks, "," ))
  slurp.Run(slurp.C, tasks...)
  slurp.Close() 
}
`))
