package glob

import (
	"reflect"
	"testing"
)

func TestDir(t *testing.T) {

	for glob, dir := range map[string]string{
		"dir/***":                    "dir",
		"dir/page/**/google/s?/page": "dir/page",
		"**": ".",
	} {

		r := Dir(glob)
		if r != dir {
			t.Fatalf("Expected %s For %s from Dir. Got %s", dir, glob, r)
		}
	}
}

func TestBase(t *testing.T) {

	for glob, dir := range map[string]string{
		"dir/***":                    "***",
		"dir/page/**/google/s?/page": "**/google/s?/page",
		"**": "**",
	} {

		r := Base(glob)
		if r != dir {
			t.Fatalf("Expected %s For %s from Base. Got %s", dir, glob, r)
		}
	}
}

func TestMatch(t *testing.T) {

	testcase := ""

	for glob, match := range map[string]bool{
		"dir/**":                      false,
		"!dir/**":                     true,
		"dir/page/**/google/s?/page":  false,
		"!dir/page/**/google/s?/page": true,
		"**": true,
		"!*": false,
	} {

		r, err := Match(glob, testcase)
		if err != nil {
			t.Fatalf("ERROR: %s  TEST: %t For %s from Match. Got %t", err, match, glob, r)
		}
		if r != match {
			t.Fatalf("Expected %t For %s from Match. Got %t", match, glob, r)
		}
	}
}

func TestFilter(t *testing.T) {

	input := []string{
		"dir/a",
		"dir/a.ext",
		"a.txt",
		"a",
		"world/hello",
		"hello/world",
		"di/page",
		"dir/a/a",
		"dir/dir/dir/file.ext",
	}

	for glob, expect := range map[string][]string{
		"dir/*":   []string{"dir/a", "dir/a.ext"},
		"!dir/**": []string{"a.txt", "a", "world/ello", "hello/world"},
	} {
		result, err := Filter(glob, input)
		if err != nil {
			t.Fatalf("ERROR: %s  TEST: %s For %s from Match. Got %s", err, expect, glob, result)
		}
		if !reflect.DeepEqual(result, expect) {
			t.Fatalf("Expected %s For %s from Match. Got %s", expect, glob, result)
		}

	}

}
