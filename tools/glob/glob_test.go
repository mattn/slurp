package glob

import "testing"

func TestDir(t *testing.T) {

	for glob, dir := range map[string]string{
		"dir/***":                    "dir",
		"dir/page/**/google/s?/page": "dir/page",
		"**": ".",
		"frontend/*.js": "frontend",
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
