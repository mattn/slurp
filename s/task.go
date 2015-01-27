package s

import (
	"fmt"
	"sync"

	"github.com/omeid/slurp/s/log"
)

type Task func(*C) error

type task struct {
	name string
	deps taskstack
	task Task

	called bool

	lock sync.Mutex
}

type taskstack map[string]*task

type taskerror struct {
	name string
	err  error
}

func (t *task) run(c *C) error {

	c = &C{c.New(fmt.Sprintf("%s: ", t.name))}

	t.lock.Lock()
	defer t.lock.Unlock()

	if t.called {
		return nil
	}
	c.Println("Starting.")

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
					c.Printf("Waiting for %s", name)
					errs <- taskerror{name, task.run(c)}
				}()
			}
		}
		wg.Wait()
	}(errs)

	var failedjobs []string

	for err := range errs {
		if err.err != nil {
			cancel <- struct{}{}
			c.Println(err.err)
			failedjobs = append(failedjobs, err.name)
		}
	}

	if failedjobs != nil {
		return fmt.Errorf("Task Canacled. Reason: Failed Dependency (%s).", failedjobs)
	}

	t.called = true
	err := t.task(c)
	if err == nil {
		c.Println("Done.")
	}

	return err
}

type Build struct {
	*C
	tasks taskstack
}

func NewBuild() *Build {
	return &Build{C: &C{log.New()}, tasks: make(taskstack)}
}

func (b *Build) Task(name string, deps []string, Task Task) {

	if _, ok := b.tasks[name]; ok {
		b.Fatalf("Duplicate task: %s", name)
	}

	Deps := make(taskstack)
	t := task{name: name, deps: Deps, task: Task}

	var ok bool

	//TODO: Circular dependency issue.
	for _, dep := range deps {
		t.deps[dep], ok = b.tasks[dep]
		if !ok {
			b.Fatalf("Missing Task %s. Required by Task %s.", dep, name)
		}
	}

	b.tasks[name] = &t
}

type Waiter interface {
	Wait()
}

func (b *Build) Run(tasks []string) Waiter {

	var wg sync.WaitGroup

	for _, name := range tasks {
		task, ok := b.tasks[name]
		if !ok {
			b.Printf("No Such Task: %s", task)
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := task.run(b.C)
			if err != nil {
				b.Println(err)
			}
		}()
	}

	return &wg
}
