// Package shell finds a command the way bash does, so a wrapper that execs it reports what typing
// the command would have: 127 when nothing is there, 126 when something is but cannot run.
package shell

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Status values of a failed lookup (bash's).
const (
	NotFound      = 127
	NotExecutable = 126
)

// Look resolves name: a name with a slash is taken as a path; else the first executable regular file
// in $PATH wins, and a non-executable match alone yields 126 (bash's "Permission denied"), as
// exec.LookPath would call it missing. status = 0 on success.
func Look(name string) (path string, status int, err error) {
	if strings.Contains(name, "/") {
		return name, check(name), describe(name)
	}
	lose := ""
	dirs := filepath.SplitList(os.Getenv("PATH"))
	if len(dirs) == 0 { // an empty or unset PATH: bash searches the current directory alone
		dirs = []string{""}
	}
	for _, dir := range dirs {
		if dir == "" {
			dir = "."
		}
		p := filepath.Join(dir, name)
		switch check(p) {
		case 0:
			return p, 0, nil
		case NotExecutable:
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() && lose == "" { //nolint:gosec // G703: the caller's own command + PATH
				lose = p
			}
		}
	}
	if lose != "" {
		return lose, NotExecutable, describe(lose)
	}
	return "", NotFound, errors.New("command not found")
}

// check = 0 for an executable regular file, else the status bash's exec of it would give.
func check(p string) int {
	fi, err := os.Stat(p) //nolint:gosec // G703: the caller's own command + PATH
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return NotFound
	case err != nil, fi.IsDir(), syscall.Access(p, 1 /* X_OK */) != nil:
		return NotExecutable
	}
	return 0
}

// describe = why exec of p fails (nil when it would not).
func describe(p string) error {
	fi, err := os.Stat(p) //nolint:gosec // G703: the caller's own command + PATH
	switch {
	case err != nil:
		return err
	case fi.IsDir():
		return errors.New("is a directory")
	case syscall.Access(p, 1) != nil:
		return errors.New("permission denied")
	}
	return nil
}
