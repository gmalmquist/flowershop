// Convenience functions for reading and writing files.
package iot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Renders a filesize in a human readable format.
//
// E.g.: 2048 becomes 2 KiB.
func HumanBytes(size uint64) string {
	thousands := []string{
		"B",
		"KiB",
		"MiB",
		"GiB",
		"TiB",
	}
	fsize := float64(size)
	var i int
	for i = 0; i < len(thousands)-1 && fsize >= 1024.0; i += 1 {
		fsize = fsize / 1024.0
	}
	if i == 0 {
		return fmt.Sprintf("%v %v", size, thousands[i])
	}
	return fmt.Sprintf("%v %v", float64(int(fsize*100))/100.0, thousands[i])
}

// Writes the given data blob to the given filepath, creating all parent directories as necessary.
//
// Overwrites any existing files.
func Powerwrite(path string, data []byte) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	parent := filepath.Dir(path)
	err = os.MkdirAll(parent, 0775)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Reads the file at the given path into a `[]byte` array, panics on errors.
//
// Useful for reading config files etc that are necessary on application startup.
func MustRead(path string) []byte {
  data, err := os.ReadFile(path)
  if err != nil {
    panic(err)
  }
  return data
}

// Returns true if a file or directory exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// Returns true if a file exists at the given path.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Returns true if a directory exists at the given path.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Returns true if the file is non-empty and contains no tabs, newlines, carriage returns, or the symbols `:"|?*/\`.
func LegalFilename(name string) bool {
	if strings.Trim(name, " .") == "" {
		return false
	}
	return !strings.ContainsAny(name, "\t\r\n<>:\"|?*/\\")
}

// Iterator which iterates over all filenames in the given folder if their corresponding
// `os.DirEntry` match the given filter.
func IterFiles(
	folder string,
	filter func(entry os.DirEntry) bool,
) func(yield func(string) bool) {
	entries, err := os.ReadDir(folder)
	return func(yield func(string) bool) {
		if err != nil {
			return
		}
		for _, e := range entries {
			name := e.Name()
			if filter != nil && !filter(e) {
				continue
			}
			if !yield(name) {
				return
			}
		}
	}
}

// Iterates over the files in the given root directory, yielding all filepaths.
//
// If file extensions are provided as varargs in `exts`, only files with the given extensions
// will be included.
func IterPaths(
  root string,
  exts ...string,
) func(yield func(string) bool) {
  allowed := map[string]bool{}
  for _, e := range exts {
    s := strings.ToLower(e)
    if s != "" && !strings.HasPrefix(s, ".") {
      s = fmt.Sprintf(".%v", s)
    }
    allowed[s] = true
  }
  accept := func(e os.DirEntry) bool {
    return len(allowed) == 0 || allowed[strings.ToLower(filepath.Ext(e.Name()))]
  }
  return func(yield func(string) bool) {
    entries, err := os.ReadDir(root)
    if err != nil {
      return
    }
    for _, e := range entries {
      if !accept(e) {
        continue
      }
      if !yield(filepath.Join(root, e.Name())) {
        return
      }
    }
  }
}

