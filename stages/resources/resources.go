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

func Name(Var string) string {
	return fmt.Sprintf(FilenameFormat, strings.ToLower(Var))
}

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

type Config struct {
	Pkg     string
	Var     string
	Declare bool
}

type Package struct {
	Config
	Files map[string]slurp.File
}

func (p *Package) Add(path string, file slurp.File) {
	p.Files[path] = file
}

func (p *Package) Build() (*bytes.Buffer, error) {
	out := new(bytes.Buffer)

	err := pkg.Execute(out, p)
	return out, err
}

func (p *Package) File() (*slurp.File, error) {

	buff, err := p.Build()
	path := Name(p.Var)
	stat := slurp.FileInfo{}
	stat.SetName(path)
	stat.SetSize(int64(buff.Len()))

	return &slurp.File{
		Reader: buff,
		Path:   path,
		Stat:   &stat,
	}, err

}

func Pack(c *slurp.C, config Config) slurp.Stage {
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
