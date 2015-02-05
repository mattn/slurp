// +build slurp

package main //Anything, even main.

import "github.com/omeid/slurp/s"

func Slurp(b *s.Build) {
	b.Task("example-task", nil, func(c *s.C) error {
		c.Println("Hello!")
		return nil
	})

	b.Task("default", []string{"example-task"}, func(c *s.C) error {
		//This task is run when slurp is called with any task parameter.
		c.Println("Hello!")
		return nil
	})
}
