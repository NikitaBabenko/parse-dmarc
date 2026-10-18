// Package pilot supplies a deliberately fictional, opt-in Maildir smoke fixture.
package pilot

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed testdata/synthetic.eml
var message []byte

const filename = "vibenest-synthetic-20261018.eml"

// Seed only writes the supplied synthetic message; it never reads a mailbox.
// A message already in new/ or cur/ is not re-created after a container replacement.
func Seed(path string) error {
	if path == "" {
		return fmt.Errorf("synthetic fixture requires MAILDIR_PATH")
	}
	for _, dir := range []string{"new", "cur", "tmp"} {
		if err := os.MkdirAll(filepath.Join(path, dir), 0700); err != nil {
			return err
		}
	}
	for _, target := range []string{filepath.Join(path, "new", filename), filepath.Join(path, "cur", filename+":2,S")} {
		if _, err := os.Stat(target); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	f, err := os.OpenFile(filepath.Join(path, "new", filename), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := f.Write(message); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
