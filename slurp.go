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

//The Slupr runner.
var install = flag.Bool("install", false, "install current build.")

var (
	gopath = os.Getenv("GOPATH")

	slurpfile = "slurp.go"

	runner string = "slurp."
	cwd    string
)

func main() {

	flag.Parse()

	if gopath == "" {
		log.Fatal("$GOPATH must be set.")
	}

	err := run(*install)
	if err != nil {
		log.Fatal(err)
	}
}

type Go struct{}

func run(install bool) error {
	path, err := generate()
	if err != nil {
		return err
	}
	//Don't forget to clean up.
	//defer os.RemoveAll(path)

	var args []string

	//if len(params) > 0 && params[0] == "init"
	get := exec.Command("go", "get")
	get.Dir = filepath.Join(path, "tmp")
	get.Stdin = os.Stdin
	get.Stdout = os.Stdout
	get.Stderr = os.Stderr

	if install {
		err := get.Run()
		if err != nil {
			return err
		}

		runnerpkg, err := filepath.Rel(filepath.Join(gopath, "src"), filepath.Join(filepath.Join(path, runner)))
		if err != nil {
			return err
		}
		args = []string{"install", runnerpkg}

	} else {
		params := flag.Args()

		if len(params) > 0 && params[0] == "init" {
			err := get.Run()
			if err != nil {
				return err
			}
		}

		args = []string{"run", filepath.Join(filepath.Join(path, runner, "main.go"))}
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

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset

	pkgs, err := parser.ParseDir(fset, cwd, nil, parser.ParseComments)
	if err != nil {
		return path, err
	}

	if len(pkgs) > 1 {
		return path, errors.New("Error: Multiple packages detected.")
	}

	for _, pkg := range pkgs {
		//This loop always runs once. I don't know of any other way to get the pkg out of pkgs
		// witout understanding the names.
		for name, f := range pkg.Files {
			f.Name.Name = "tmp" //Change package name
			if filepath.Base(name) == slurpfile {
				f.Comments = []*ast.CommentGroup{} //Remove comments
			}

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

	tmp, err = filepath.Rel(filepath.Join(gopath, "src"), path)
	if err != nil {
		return path, err
	}

	err = runnerSrc.Execute(file, tmp) //This should never fail, see MustParse.
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

  "github.com/omeid/slurp/s"

  client "{{ . }}/tmp"
)

func main() {

  flag.Parse()

  slurp := s.NewBuild()

  client.Slurp(slurp)

  tasks := flag.Args()
  if len(tasks) == 0 {
	tasks = []string{"default"}
  }

  slurp.Printf("Running: %s", strings.Join(tasks, ","))
  slurp.Run(tasks).Wait()
  slurp.Println("Finished.")
}
`))
