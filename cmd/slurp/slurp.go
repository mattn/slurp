package main

import (
	"errors"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	gopath    = os.Getenv("GOPATH")
	gopathsrc = filepath.Join(gopath, "src")
	cwd       string

	build     = flag.Bool("build", false, "build the current build as slurp-bin")
	install   = flag.Bool("install", false, "install current slurp.Go as slurp.PKG.")
	bare      = flag.Bool("bare", false, "Run/Install the slurp.go file without any other files.")
	slurpfile = flag.String("slurpfile", "slurp.go", "The file that includes the Slurp(*s.Build) function, use by -bare")

	keep = flag.Bool("keep", false, "keep the generated source under $GOPATH/src/slurp-run-*")
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
	runnerpath := filepath.Join(path, "runner")
	if err != nil {
		return err
	}

	//Don't forget to clean up.
	if !*keep {
		//	defer os.RemoveAll(path)
	}

	//if len(params) > 0 && params[0] == "init"
	get := exec.Command("go", "get", "-tags=slurp", "-v")
	get.Dir = filepath.Join(path, "tmp")
	get.Stdin = os.Stdin
	get.Stdout = os.Stdout
	get.Stderr = os.Stderr

	if *build || *install {
		err := get.Run()
		if err != nil {
			return err
		}
	}

	var args []string

	if *build {
		args = []string{"build", "-tags=slurp", "-o=slurp-bin", runnerpath}

	} else if *install {
		args = []string{"install", "-tags=slurp", runnerpath}

	} else {
		params := flag.Args()

		if len(params) > 0 && params[0] == "init" {
			err := get.Run()
			if err != nil {
				return err
			}
		}

		args = []string{"run", "-tags=slurp", filepath.Join(runnerpath, "main.go")}
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

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	//The target package import path.
	pkgpath, err := filepath.Rel(gopathsrc, cwd)
	if err != nil {
		return "", err
	}

	if base := filepath.Base(pkgpath); base == "." || base == ".." {
		return "", errors.New("forbidden path. Your CWD must be under $GOPATH/src.")
	}

	//build our package path.
	path := filepath.Join(gopathsrc, "slurp", pkgpath)

	//Clean it up.
	os.RemoveAll(path)

	//log.Println("Creating temporary build path...", path)

	//Create the target package directory.
	tmp := filepath.Join(path, "tmp")
	err = os.MkdirAll(tmp, 0700)
	if err != nil {
		return path, err
	}

	//Create the runner package directory.
	runnerpath := filepath.Join(path, "runner")
	err = os.Mkdir(runnerpath, 0700)
	if err != nil {
		return path, err
	}

	//TODO, copy [*.go !_test.go] files into tmp first,
	// this would allow slurp to work for broken packages
	// with "-bare" as the package files will be excluded.
	fset := token.NewFileSet() // positions are relative to fset

	var pkgs map[string]*ast.Package

	//log.Printf("Parsing %s...", pkgpath)

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

	//log.Println("Generating the runner...")
	file, err := os.Create(filepath.Join(runnerpath, "main.go"))
	if err != nil {
		return path, err
	}

	tmp, err = filepath.Rel(filepath.Join(gopath, "src"), path)
	if err != nil {
		return path, err
	}

	err = runnerSrc.Execute(file, filepath.ToSlash(tmp))
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
