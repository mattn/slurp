package slurp

import (
	"sync"

	"github.com/omeid/slurp/log"
)

type Waiter interface {
	Wait()
}

type Build struct {
	*C
	tasks taskstack

	cleanups []func()
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

	for _, dep := range deps {
		d, ok := b.tasks[dep]
		if !ok {
			b.Fatalf("Missing Task %s. Required by Task %s.", dep, name)
		}
		_, ok = d.deps[name]
		if ok {
			b.Fatalf("Circular dependency %s requies %s and around.", d.name, name)
		}

		t.deps[dep] = d
	}

	b.tasks[name] = &t
}

func (b *Build) Run(c *C, tasks ...string) {
	b.Start(c, tasks...).Wait()
}

func (b *Build) Start(c *C, tasks ...string) Waiter {
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
			err := task.run(c)
			if err != nil {
				b.Println(err)
			}
		}()
	}

	return &wg
}

// Register a function to be called when gulp exists.
func (b *Build) Defer(fn func()) {
	b.cleanups = append(b.cleanups, fn)
}

// Helper function. waits forever.
func (b Build) Wait() {
	<-make(chan struct{})
}

func (b *Build) Close() {
	for _, cleanup := range b.cleanups {
		cleanup()
	}
}
