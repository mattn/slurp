// Experimental resource embedding stage for Slurp.

package resources

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/omeid/slurp"
)

var FilenameFormat = "%s_resource.go"

//Create a new Package.
func New() *Package {
	return &Package{
		Config: Config{
			Pkg:     "resources",
			Var:     "FS",
			Declare: true,
		},
		Files: make(map[string]slurp.File),
	}
}

//Configuration defines some details about the output Go file.
// Pkg      the package name to use.
// Var      the variable name to assign the file system to.
// Declare  dictates whatever there should be a defintion for the variable
//          in the output file or not, it will use the type http.FileSystem.
type Config struct {
	Pkg     string
	Var     string
	Declare bool
}

type Package struct {
	Config
	Files map[string]slurp.File
}

//Add a file to the package at the give path.
func (p *Package) Add(path string, file slurp.File) {
	p.Files[path] = file
}


//Build the package
func (p *Package) Build() (*bytes.Buffer, error) {
	out := new(bytes.Buffer)
	return out, pkg.Execute(out, p)
}

// Returns the build as a *slurp.File, you do not need to
// call build yourself.
func (p *Package) File() (*slurp.File, error) {

	buff, err := p.Build()
	path := fmt.Sprintf(FilenameFormat, strings.ToLower(p.Var))
	stat := slurp.FileInfo{}
	stat.SetName(path)
	stat.SetSize(int64(buff.Len()))

	return &slurp.File{
		Reader: buff,
		Path:   path,
		Stat:   &stat,
	}, err

}

// A build stage creates a new Package and adds all the files coming through the channel to
// the package and returns the result of build as a File on the output channel.
func Stage(c *slurp.C, config Config) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		res := New()
		res.Config = config
		for file := range in {
			path, _ := filepath.Rel(file.Dir, file.Path)
			res.Add(path, file)
			c.Printf("Adding %s.\n", path)
			defer file.Close() //Close files AFTER we have build our package.
		}

		f, err := res.File()
		if err != nil {
			c.Println(err)
			return
		}
		out <- *f
	}
}
