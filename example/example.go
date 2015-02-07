// +build slurp

package main //Anything, even main.

import "github.com/omeid/slurp"

func Slurp(b *slurp.Build) {
	b.Task("example-task", nil, func(c *slurp.C) error {
		c.Println("Hello!")
		return nil
	})

	b.Task("default", []string{"example-task"}, func(c *slurp.C) error {
		//This task is run when slurp is called with any task parameter.
		c.Println("Hello!")
		return nil
	})
}
