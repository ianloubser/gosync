package main

import "testing"

func TestCanonicalFileKey(t *testing.T) {

	path, _ := getCanonicalFileKey("C:/Users/test/some/dir/to/file.txt")

	if path != "/Users/test/some/dir/to/file.txt" {
		t.Fatalf("expected volume to be stripped and unix path format")
	}
}
