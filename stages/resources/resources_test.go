package resources_test

import (
	"testing"

	"github.com/omeid/slurp/resources"
)

func TestNew(t *testing.T) {
	pkg := resources.New()

	err := pkg.AddFile("testfile.txt")
	if err != nil {
		t.Fatal(err)
	}

	out, err := pkg.Build()
	if err != nil {
		t.Fatal(err)
	}
	_ = out
	//	out.WriteTo(os.Stdout)
}
