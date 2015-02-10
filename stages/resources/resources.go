package resources

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/omeid/slurp"
	"github.com/omeid/slurp/stages/fs"
)

var FilenameFormat = "%s_resource.go"

func Name(Var string) string {
	return fmt.Sprintf(FilenameFormat, strings.ToLower(Var))
}

func New() *Package {
	return &Package{
		Pkg:   "resources",
		Var:   "FS",
		Files: make(map[string]slurp.File),
	}
}

type Package struct {
	Pkg   string
	Var   string
	Files map[string]slurp.File
}

func (p *Package) Add(file slurp.File) {
	p.Files[file.Path] = file
}

func (p *Package) AddFile(path string) error {
	f, err := fs.Read(path)
	if err != nil {
		return err
	}

	p.Add(*f)
	return nil
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

func Pack(c *slurp.C, Pkg string, Var string) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		res := New()
		res.Pkg = Pkg
		res.Var = Var
		for file := range in {
			res.Add(file)
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
