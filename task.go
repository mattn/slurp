package m

import (
	"fmt"
	"log"
	"sync"
)

type Task func() error

type task struct {
	deps taskstack
	task Task
}

type taskstack map[string]*task

type taskerror struct {
	name string
	err  error
}

func (t *task) run() error {

	done := make(chan taskerror)
	cancel := make(chan struct{})

	for name, task := range t.deps {
		select {
		case <-cancel:
			break
		default:
			go func() {
				done <- taskerror{name, task.run()}
			}()
		}
	}

	var failedjobs []string

	for te := range done {
		if te.err != nil {
			if cancel != nil {
				close(cancel)
			}

			log.Println(te.err)
			failedjobs = append(failedjobs, te.name)
		}
	}

	if failedjobs != nil {
		return fmt.Errorf("Task Canacled. Reason: Failed Dependency (%s).", failedjobs)
	}

	return t.task()
}

type Build struct {
	tasks taskstack
}

func NewBuild() *Build {
	return &Build{tasks: make(taskstack)}
}

func (b *Build) Task(name string, deps []string, Task Task) {
	Deps := make(taskstack)
	t := task{deps: Deps, task: Task}

	var ok bool

	//TODO: Circular dependency issue.
	for _, dep := range deps {
		t.deps[dep], ok = b.tasks[dep]
		if !ok {
			log.Fatalf("Missing Task %s. Required by Task %s.")
		}
	}

	b.tasks[name] = &t
}

func (b *Build) Run(tasks ...string) {

	var wg sync.WaitGroup

	for _, name := range tasks {
		task, ok := b.tasks[name]

		if !ok {
			log.Fatalf("No Such Task: %s", task)
		}
		wg.Add(1)
		task.run()
	}

	wg.Wait()
}
