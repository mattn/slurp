package s

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

	errs := make(chan taskerror)
	cancel := make(chan struct{}, len(t.deps))
	var wg sync.WaitGroup
	go func(done chan taskerror) {
		defer close(errs)
		for name, task := range t.deps {
			select {
			case <-cancel:
				break
			default:
				wg.Add(1)
				go func() {
					defer wg.Done()
					errs <- taskerror{name, task.run()}
				}()
			}
		}
		wg.Wait()
	}(errs)

	var failedjobs []string

	for err := range errs {
		if err.err != nil {
			cancel <- struct{}{}
			log.Println(err.err)
			failedjobs = append(failedjobs, err.name)
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

type Waiter interface {
	Wait()
}

func (b *Build) Run(tasks ...string) Waiter {

	var wg sync.WaitGroup

	for _, name := range tasks {
		task, ok := b.tasks[name]
		if !ok {
			log.Printf("No Such Task: %s", task)
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := task.run()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	return &wg
}
