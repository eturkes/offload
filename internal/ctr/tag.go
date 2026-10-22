package ctr

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
)

// Repo = the image repository every tag lives in.
const Repo = "localhost/offload-ctr"

// Tag = Repo:<16 hex of sha256 over the context tree> (K4): each entry below dir in path-byte
// order as kind, permission bits (COPY keeps them), path + content (file bytes, symlink target),
// length-framed. Owners + mtimes never count (COPY sets root; --timestamp 0 fixes times), so every
// host with the same tree gets the same tag.
func Tag(dir string) (string, error) {
	type entry struct{ rel, kind, perm, content string }
	var entries []entry
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("%s: no stat data", path)
		}
		e := entry{rel: rel, perm: strconv.FormatUint(uint64(st.Mode&0o7777), 8)}
		switch t := d.Type(); {
		case t.IsDir():
			e.kind = "d"
		case t&fs.ModeSymlink != 0:
			e.kind, e.perm = "l", ""
			e.content, err = os.Readlink(path)
		case t.IsRegular():
			e.kind = "f"
			var b []byte
			b, err = os.ReadFile(path) //nolint:gosec // G304: the project's own build context
			e.content = string(b)
		default:
			return fmt.Errorf("%s: unsupported file type %s", path, t)
		}
		entries = append(entries, e)
		return err
	})
	if err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })
	h := sha256.New()
	for _, e := range entries {
		for _, field := range []string{e.kind, e.perm, e.rel, e.content} {
			h.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
		}
	}
	return Repo + ":" + hex.EncodeToString(h.Sum(nil))[:16], nil
}
