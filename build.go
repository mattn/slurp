package slurp

import (
	"sync"

	"github.com/omeid/slurp/log"
)

// Waiter interface implementsion Wait() function.
type Waiter interface {
	Wait()
}


// Build is a simple build harness that you can register tasks and their 
// dependencies and then run them. You usually don't need to create your
// own Build and instead use the one passed by Slurp runner.
type Build struct {
	*C
	tasks taskstack

	cleanups []func()
}


func NewBuild() *Build {
	return &Build{C: &C{log.New()}, tasks: make(taskstack)}
}


// Register a task and it's dependencies.
// When running the task, the dependencies will be run in parallel.
// Circular Dependencies are not allowed and will result into error.
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

// Run Starts a task and waits for it to finish.
func (b *Build) Run(c *C, tasks ...string) {
	b.Start(c, tasks...).Wait()
}

// Start a task but doesn't wait for it to finish.
func (b *Build) Start(c *C, tasks ...string) Waiter {
	var wg sync.WaitGroup

	for _, name := range tasks {
		task, ok := b.tasks[name]
		if !ok {
			b.Printf("No Such Task: %s", name)
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

// Register a function to be called when Slurp exists.
func (b *Build) Defer(fn func()) {
	b.cleanups = append(b.cleanups, fn)
}

// Helper function. waits forever.
func (b Build) Wait() {
	<-make(chan struct{})
}

//Close a build, it will call all the cleanup functions.
func (b *Build) Close() {
	for _, cleanup := range b.cleanups {
		cleanup()
	}
}
