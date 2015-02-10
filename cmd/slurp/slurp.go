package main

import (
	"errors"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var (
	gopath        = os.Getenv("GOPATH")
	runner string = "slurp."
	cwd    string

	build = flag.Bool("build", false, "build the current build as slurp-bin")
	install   = flag.Bool("install", false, "install current slurp.Go as slurp.PKG.")
	bare      = flag.Bool("bare", false, "Run/Install the slurp.go file without any other files.")
	slurpfile = flag.String("slurpfile", "slurp.go", "The file that includes the Slurp(*s.Build) function, use by -bare")
)

func main() {

	flag.Parse()

	if gopath == "" {
		log.Fatal("$GOPATH must be set.")
	}

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path, err := generate()
	if err != nil {
		return err
	}

	//Don't forget to clean up.
	defer os.RemoveAll(path)

	var args []string

	//if len(params) > 0 && params[0] == "init"
	get := exec.Command("go", "get", "-tags=slurp", "-v")
	get.Dir = filepath.Join(path, "tmp")
	get.Stdin = os.Stdin
	get.Stdout = os.Stdout
	get.Stderr = os.Stderr

	if *build {
		err := get.Run()
		if err != nil {
			return err
		}

		runnerpkg, err := filepath.Rel(filepath.Join(gopath, "src"), filepath.Join(filepath.Join(path, runner)))
		if err != nil {
			return err
		}
		args = []string{"build", "-tags=slurp", "-o=slurp-bin", runnerpkg}

	} else if *install {
		err := get.Run()
		if err != nil {
			return err
		}

		runnerpkg, err := filepath.Rel(filepath.Join(gopath, "src"), filepath.Join(filepath.Join(path, runner)))
		if err != nil {
			return err
		}
		args = []string{"install", "-tags=slurp", runnerpkg}

	} else {
		params := flag.Args()

		if len(params) > 0 && params[0] == "init" {
			err := get.Run()
			if err != nil {
				return err
			}
		}

		args = []string{"run", "-tags=slurp", filepath.Join(filepath.Join(path, runner, "main.go"))}
		args = append(args, params...)
	}

	cmd := exec.Command("go", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

func generate() (string, error) {

	//Let's grab a temp folder.
	path, err := ioutil.TempDir(filepath.Join(gopath, "src"), "slurp-run-")
	if err != nil {
		return "", err
	}

	tmp := filepath.Join(path, "tmp")
	err = os.Mkdir(tmp, 0700)
	if err != nil {
		return path, err
	}

	cwd, err = os.Getwd()
	if err != nil {
		return path, err
	}

	runner = runner + filepath.Base(cwd)
	runnerpkg := filepath.Join(path, runner)
	err = os.Mkdir(runnerpkg, 0700)
	if err != nil {
		return path, err
	}

	//TODO, copy [*.go !_test.go] files into tmp first,
	// this would allow slurp to work for broken packages
	// with "-bare" as the package files will be excluded.
	fset := token.NewFileSet() // positions are relative to fset

	var pkgs map[string]*ast.Package

	if *bare {
		pkgs = make(map[string]*ast.Package)
		src, err := parser.ParseFile(fset, *slurpfile, nil, parser.ParseComments)
		if err != nil {
			return path, err
		}
		pkgs[src.Name.Name] = &ast.Package{
			Name:  src.Name.Name,
			Files: map[string]*ast.File{filepath.Join(cwd, *slurpfile): src},
		}
	} else {
		pkgs, err = parser.ParseDir(fset, cwd, nil, parser.ParseComments)
		if err != nil {
			return path, err
		}
	}

	if len(pkgs) > 1 {
		return path, errors.New("Error: Multiple packages detected.")
	}

	for _, pkg := range pkgs {
		//This loop always runs once. I don't know of any other way to get the pkg out of pkgs
		// witout understanding the names.
		for name, f := range pkg.Files {
			f.Name.Name = "tmp" //Change package name

			name, err = filepath.Rel(cwd, name)
			if err != nil {
				//Should never get error. But just incase.
				return path, err
			}
			err = writeFileSet(filepath.Join(tmp, name), fset, f)
			if err != nil {
				return path, err
			}
		}
	}

	file, err := os.Create(filepath.Join(runnerpkg, "main.go"))
	if err != nil {
		return path, err
	}

	tmp, err = filepath.Rel(filepath.Join(gopath, "src"), path)
	if err != nil {
		return path, err
	}

	err = runnerSrc.Execute(file, tmp)
	if err != nil {
		return path, err
	}

	err = file.Close()
	if err != nil {
		return path, err
	}

	return path, nil

}

func writeFileSet(filepath string, fset *token.FileSet, node interface{}) error {
	// Print the modified AST.
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return format.Node(file, fset, node)
}

var runnerSrc = template.Must(template.New("main").Parse(`
package main

import (
  "flag"
  "strings"
  "os"
  "os/signal"

  "github.com/omeid/slurp"

  client "{{ . }}/tmp"
)

func main() {

  flag.Parse()

  interrupts := make(chan os.Signal, 1)
  signal.Notify(interrupts, os.Interrupt)

  slurp := slurp.NewBuild()

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
