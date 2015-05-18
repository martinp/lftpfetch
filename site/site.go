package site

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/martinp/lftpfetch/cmd"
)

type Client struct {
	LftpGetCmd string
	LftpPath   string
}

type Site struct {
	Client
	Name         string
	Dir          string
	MaxAge       string
	maxAge       time.Duration
	Patterns     []string
	patterns     []*regexp.Regexp
	Filters      []string
	filters      []*regexp.Regexp
	SkipSymlinks bool
	ParseTVShow  bool
	LocalDir     string
	localDir     *template.Template
}

func (s *Site) ListCmd() cmd.Lftp {
	script := "cls --date --time-style='%F %T %z %Z' " + s.Dir
	return cmd.Lftp{
		Path:   s.LftpPath,
		Script: script,
		Site:   s.Name,
	}
}

func (s *Site) DirList() ([]Dir, error) {
	listCmd := s.ListCmd()
	cmd := listCmd.Cmd()
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	dirs := []Dir{}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t\r\n")
		if len(line) == 0 {
			continue
		}
		dir, err := ParseDir(line)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, dir)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return dirs, nil
}

func (s *Site) Queue(dirs []Dir) (*Queue, error) {
	queue := Queue{Site: *s}
	if err := queue.Process(dirs); err != nil {
		return nil, err
	}
	return &queue, nil
}
