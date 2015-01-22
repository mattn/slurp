package main

import (
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

var runner = template.Must(template.New("main").Parse(`
package main

import (
  "fmt"

  client "{{ . }}"
)

func init() {
  fmt.Println("Starting...")
}

func main() {
  fmt.Println("Running....")
  client.Slurp()
  fmt.Println("End.")
}`))

var (
	slurpfile = "slurp.go"
	cwd       string
	gopath    = os.Getenv("GOPATH")
)

func main() {

	if gopath == "" {
		log.Fatal("$GOPATH must be set.")
	}

	err, path := generate()
	if err != nil {
		log.Fatal(err)
	}
	//Don't forget to clean up.
	defer os.RemoveAll(path)

	cmd := exec.Command("go", "run",
		filepath.Join(path, "main.go"),
		filepath.Join(path, "slurp.go"),
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done.")
}

func generate() (error, string) {

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, slurpfile, nil, parser.ParseComments)
	if err != nil {
		return err, ""
	}

	f.Name.Name = "main"               //Change package name
	f.Comments = []*ast.CommentGroup{} //Remove comments

	tmp, err := ioutil.TempDir(os.TempDir(), "slurp-")
	if err != nil {
		return err, ""
	}

	err = writeFileSet(filepath.Join(tmp, "slurp.go"), fset, f)
	if err != nil {
		return err, tmp
	}

	file, err := os.Create(filepath.Join(tmp, "main.go"))

	cwd, err = os.Getwd()
	if err != nil {
		return err, tmp
	}

	path, err := filepath.Rel(filepath.Join(gopath, "src"), cwd)
	if err != nil {
		return err, tmp
	}

	err = runner.Execute(file, path) //This should never fail, see MustParse.
	err = file.Close()

	if err != nil {
		return err, tmp
	}

	return nil, tmp

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
