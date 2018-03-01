package main

import (
	"encoding/hex"
	"testing"
)

func TestCanonicalFileKey(t *testing.T) {

	path1, _ := getCanonicalFileKey("C:/Users/test/some/dir/to/file.txt")
	path2, _ := getCanonicalFileKey("/usr/local/lib/awe/some/file.txt")
	path3, _ := getCanonicalFileKey("C:\\Users\\test\\some\\dir\\to\\file.txt")

	if path1 != "/Users/test/some/dir/to/file.txt" {
		t.Fatalf("expected volume to be stripped and unix path format")
	}

	if path2 != "/usr/local/lib/awe/some/file.txt" {
		t.Fatalf("expected unix path to remain the same")
	}

	if path3 != "/Users/test/some/dir/to/file.txt" {
		t.Fatalf("expected volume to be stripped slashes correct for unix path format")
	}
}

func TestGetFileHash(t *testing.T) {
	hash, _ := getMD5("./sample_hash_file.txt")
	md5Hash := hex.EncodeToString(hash.MD5)

	if md5Hash != "efd1014031223f2dad019786443ec318" {
		t.Fatalf("Failed MD5 hash comparison: %s", md5Hash)
	}
}
