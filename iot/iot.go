package iot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func MustRead(path string) []byte {
  data, err := os.ReadFile(path)
  if err != nil {
    panic(err)
  }
  return data
}


func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func LegalFilename(name string) bool {
	if strings.Trim(name, " .") == "" {
		return false
	}
	return !strings.ContainsAny(name, "\t\r\n<>:\"|?*/\\")
}

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

