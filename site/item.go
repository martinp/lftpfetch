package site

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinp/lftpq/parser"
)

type Item struct {
	Dir
	LocalDir string
	Transfer bool
	Reason   string
	Media    interface{}
	*Queue
}

func (i *Item) String() string {
	return fmt.Sprintf("Path=%q LocalDir=%q Transfer=%t Reason=%q", i.Path, i.LocalDir, i.Transfer, i.Reason)
}

func (i *Item) IsDstDirEmpty() bool {
	dstDir := i.LocalDir
	// When LocalDir has a trailing slash, the actual dstDir will be a directory inside LocalDir
	// (same behaviour as rsync)
	if strings.HasSuffix(dstDir, string(os.PathSeparator)) {
		dstDir = filepath.Join(dstDir, i.Dir.Base())
	}
	dirs, _ := ioutil.ReadDir(dstDir)
	return len(dirs) == 0
}

func (i *Item) ShowEqual(o Item) bool {
	a, ok := i.Media.(parser.Show)
	if !ok {
		return false
	}
	b, ok := o.Media.(parser.Show)
	if !ok {
		return false
	}
	return a.Equal(b)
}

func (i *Item) MovieEqual(o Item) bool {
	a, ok := i.Media.(parser.Movie)
	if !ok {
		return false
	}
	b, ok := o.Media.(parser.Movie)
	if !ok {
		return false
	}
	return a.Equal(b)
}

func (i *Item) MediaEqual(o Item) bool {
	return i.ShowEqual(o) || i.MovieEqual(o)
}

func (i *Item) Weight() int {
	for _i, p := range i.Queue.priorities {
		if i.Dir.Match(p) {
			return len(i.Queue.priorities) - _i
		}
	}
	return 0
}

func (i *Item) parseMedia() (interface{}, error) {
	switch i.Queue.Parser {
	case "show":
		show, err := i.Dir.Show()
		if err != nil {
			return nil, err
		}
		return show, nil
	case "movie":
		movie, err := i.Dir.Movie()
		if err != nil {
			return nil, err
		}
		return movie, nil
	}
	return nil, nil
}

func (i *Item) parseLocalDir() (string, error) {
	var b bytes.Buffer
	if err := i.Queue.localDir.Execute(&b, i.Media); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (i *Item) setMetadata() {
	i.LocalDir = i.Queue.LocalDir
	if i.Queue.Parser == "" {
		return
	}

	m, err := i.parseMedia()
	if err != nil {
		i.Transfer = false
		i.Reason = err.Error()
		return
	}
	i.Media = m

	d, err := i.parseLocalDir()
	if err != nil {
		i.Transfer = false
		i.Reason = err.Error()
		return
	}
	i.LocalDir = d
}

func newItem(q *Queue, d Dir) Item {
	item := Item{Queue: q, Dir: d, Reason: "no match"}
	item.setMetadata()
	return item
}