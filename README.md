> Heads up! This is pre-release software.

# Slurp 
Building with Go, easier than a slurp.

Slurp is a [Gulp.js](http://gulpjs.com/) inspired build framework and toolkit designed with idiomatic Go [Pipelines](http://blog.golang.org/pipelines) and following principles: 

- Convention over configuration
- Explicit is better than implicit.
- Do one thing. Do it well.
- ...


### Why?
> The tale of Gulp, Go, Go Templates, CSS, CoffeeScript, and minifiaction and building assets should go here.


##### I am interested, how it works?

Slurp is made of two integral parts:

### 1. The Framework 

The slurp framework provides a task harness that you can register tasks and dependencies, you can then run these tasks with slurp runner.

A task is any function that accepts a pointer to `s.C` (Slurp Context) and returns an error.  
The Context provides helpful logging functions. _it may be extended in the future_.

```go
b.Task("example-task", []string{"list", "of", "dependency", "tasks"},

  func(c *slurp.C) error {
    c.Println("Hello from example-task!")
  },

)
```

Following the Convention Over Configuration paradigm, slurps provides you with a collection of nimble tools to instrument a pipeline.

A pipeline is created by a source _stage_ and typically piped to subsequent _transformation_ stages and a final _destitution_ stage.

Currently Slurp provides two source stages `slurp/stages/fs` and `slurp/stages/web` that provide access to file-system and http source respectively.

```go
b.Task("example-task-with-pipeline", nil , func(c *slurp.C) error {
    //Read .tpl files from frontend/template.
    return fs.Src(c, "frontend/template/*.tpl").Pipe(
      //Compile them.
      template.HTML(c, TemplateData),
      //Write the result to disk.
      fs.Dest(c, "./public"),
    ).Wait() //Wait for all to finish.
})
```

```go
// Download deps.
b.Task("deps", nil, func(c *slurp.C) error {
    return web.Get(c,
      "https://github.com/twbs/bootstrap/archive/v3.3.2.zip",
      "https://github.com/FortAwesome/Font-Awesome/archive/v4.3.0.zip",
    ).Pipe(
      archive.Unzip(c),
      fs.Dest(c, "./frontend/libs/"),
    ).Wait()

})
```


### 2. The Runner (cmd/slurp)

This is a cli tool that runs and help you compile your builders. It is go get’able and you can install with:

```bash
 $ go get github.com/omeid/slurp               # get it.
 $ go install github.com/omeid/slurp/cmd/slurp # install it.
```

Slurp uses the Slurp build tag. That is, it passes `-tags=slurp` to go tooling when building or running your project,
this allows decoupling of build and project code. This means you can use Go tools just like you're used to, even if your
project has a slurp file.

Somewhat similar to `go test` Slurp expects a `Slurp(*b.Build)` function from your project, this is typically put in a file with the `// +build slurp` build tag.

`cat slurp.go`
```go
// +build slurp

package main //Anything, even main.

import "github.com/omeid/slurp"

func Slurp(b *slurp.Build) {
	b.Task("example-task", nil, func(c *slurp.C) error {
		c.Println("Hello!")
		return nil
	})

	b.Task("default", []string{"example-task"}, func(c *slurp.C) error {
		//This task is run when slurp is called without any task arguments.
		c.Println("Hello!")
		return nil
	})
}
```
```bash
$ slurp 
09:12:48 Running: default
09:12:48 default: Starting.
09:12:48 default: Waiting for example-task
09:12:48 default: example-task: Starting.
09:12:48 default: example-task: Hello!
09:12:48 default: example-task: Done.
09:12:48 default: Hello!
09:12:48 default: Done.

$ slurp example-task
09:13:02 Running: example-task
09:13:02 example-task: Starting.
09:13:02 example-task: Hello!
09:13:02 example-task: Done.
```
### Contributing

Please see [Contributing](CONTRIBUTING.md)


### TODO

### LICENSE
The MIT License (MIT)
Copyright (c) 2014 omeid <public@omeid.me>

See [LICENSE](LICENSE).
