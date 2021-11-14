package pkg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/morganhein/autostart.sh/pkg/T"
	"github.com/morganhein/autostart.sh/pkg/io"
)

// InsureSymlinks For every file in $source, make sure it is symlinked into $target.
// If a file already exists in $target with the same name, and is *not* a symlink from origin,
// then ask to a: move the file b: delete it, c: merge it

type FileMismatchIssue string

const (
	MissingFromTarget FileMismatchIssue = "missing from target"
	MissingFromSource FileMismatchIssue = "missing from source"
	FileCollision     FileMismatchIssue = "file collision"
)

type Mismatch struct {
	From  string
	To    string
	Issue FileMismatchIssue
}

type walker struct {
	ctx        context.Context
	fs         io.Filesystem
	baseSource string // always the $home_source or $root_source
	baseTarget string // always the $home_target or $root_target
	issues     []Mismatch
	ignores    []string // list of filenames or regular expressions to ignore, to be added
	log        io.Logger
	linkDirs   bool // if enabled, don't link individual files, symlink entire directories
}

func (w *walker) isIgnored(pathname string) bool {
	for _, v := range w.ignores {
		if pathname == v {
			return true
		}
	}
	return false
}

func (w *walker) GoWalkerSourceToTarget(pathName string, dir *godirwalk.Dirent) error {
	//keep things sane and scan the correct folders
	if filepath.Clean(pathName) == filepath.Clean(w.baseSource) {
		return godirwalk.SkipThis
	}
	if filepath.Clean(pathName) == filepath.Clean(w.baseTarget) {
		return nil
	}
	if w.isIgnored(pathName) {
		w.log.Debugf("skipping %v", pathName)
		return godirwalk.SkipThis
	}
	if dir.IsDir() && !w.linkDirs {
		return nil
	}
	if !dir.IsDir() && w.linkDirs {
		return godirwalk.SkipThis
	}

	return w.sourceToTargetHelper(pathName)
}

func (w *walker) sourceToTargetHelper(pathName string) error {
	//get relative path
	relativePath := strings.TrimPrefix(pathName, w.baseTarget)
	sourcePath := filepath.Join(w.baseSource, relativePath)
	//check if this file also exists in target
	_, err := w.fs.Stat(sourcePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist, symlink and return
		w.issues = append(w.issues, Mismatch{
			From:  pathName,
			To:    sourcePath,
			Issue: MissingFromSource,
		})
		return nil
	}
	if err != nil {
		return T.Log(err)
	}

	//check if path is already symlinking to sourcePath
	alreadyLinked, err := w.fs.IsSymlinkTo(pathName, sourcePath)
	if err != nil {
		return T.Log(err)
	}
	if alreadyLinked {
		return nil
	}

	//a match exists, but is not a symlink to the correct location
	w.issues = append(w.issues, Mismatch{
		From:  pathName,
		To:    sourcePath,
		Issue: FileCollision,
	})
	return nil
}

func (w *walker) GoWalkerTargetToSource(pathName string, dir *godirwalk.Dirent) error {
	//skip this if this is the source repo
	if pathName == w.baseSource {
		return godirwalk.SkipThis
	}
	if w.isIgnored(pathName) {
		return godirwalk.SkipThis
	}
	if dir.IsDir() {
		if w.linkDirs {
			//if it's a di
			return godirwalk.SkipThis
		}
		return nil
	}
	//get relative path
	relativePath := strings.TrimPrefix(pathName, w.baseTarget)
	sourcePath := filepath.Join(w.baseSource, relativePath)
	//check if this file also exists in target
	_, err := w.fs.Stat(sourcePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist, symlink and return
		w.issues = append(w.issues, Mismatch{
			From:  pathName,
			To:    sourcePath,
			Issue: MissingFromSource,
		})
		return nil
	}
	if err != nil {
		return T.Log(err)
	}

	//check if path is already symlinking to sourcePath
	alreadyLinked, err := w.fs.IsSymlinkTo(pathName, sourcePath)
	if err != nil {
		return T.Log(err)
	}
	if alreadyLinked {
		return nil
	}

	//a match exists, but the symlink is to the wrong location
	w.issues = append(w.issues, Mismatch{
		From:  pathName,
		To:    sourcePath,
		Issue: FileCollision,
	})
	return nil
}

/* stdlib implementations

type Direction string

const (
	FromConfig Direction = "from config" //checking that everything in the config repo is linked in target
	ToConfig   Direction = "to config"   //checking that everything in target exists in config repo
)


func (w *walker) WalkSourceToTarget(path string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	//get relative path
	relativePath := strings.TrimPrefix(path, w.baseSource)
	targetPath := filepath.Join(w.baseTarget, relativePath)
	//check if this file also exists in target
	_, err = w.fs.Stat(targetPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist, symlink and return
		w.issues = append(w.issues, Mismatch{
			From:  path,
			To:    targetPath,
			Issue: MissingFromTarget,
		})
		return nil
	}
	if err != nil {
		return err
	}

	//check if target is a symlink as well, notice the source/target switches here
	alreadyLinked, err := w.fs.IsSymlinkTo(targetPath, path)
	if err != nil {
		return err
	}
	if alreadyLinked {
		return nil
	}

	//a file already exists at the target that is not a symlink, handle
	w.issues = append(w.issues, Mismatch{
		From:  path,
		To:    targetPath,
		Issue: FileCollision,
	})
	return nil
}

func (w *walker) WalkTargetToSource(pathName string, dir fs.DirEntry, err error) error {
	if dir.IsDir() {
		return nil
	}
	//get relative path
	relativePath := strings.TrimPrefix(pathName, w.baseTarget)
	sourcePath := filepath.Join(w.baseSource, relativePath)
	//check if this file also exists in target
	_, err = w.fs.Stat(sourcePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist, symlink and return
		w.issues = append(w.issues, Mismatch{
			From:  pathName,
			To:    sourcePath,
			Issue: MissingFromSource,
		})
		return nil
	}
	if err != nil {
		return err
	}

	//check if path is already symlinking to sourcePath
	alreadyLinked, err := w.fs.IsSymlinkTo(pathName, sourcePath)
	if err != nil {
		return err
	}
	if alreadyLinked {
		return nil
	}

	//a match exists, but the symlink is to the wrong location
	w.issues = append(w.issues, Mismatch{
		From:  pathName,
		To:    sourcePath,
		Issue: FileCollision,
	})
	return nil
}
*/
