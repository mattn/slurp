package s

import (
	"fmt"
	"sync"
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
	go func(errs chan taskerror) {
		defer close(errs)
		for name, t := range t.deps {
			select {
			case <-cancel:
				break
			default:

				wg.Add(1)
				go func(t *task) {
					defer wg.Done()
					c.Printf("Waiting for %s", t.name)
					errs <- taskerror{name, t.run(c)}
				}(t)
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
