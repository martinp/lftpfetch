package site

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	LftpGetCmd string
	LftpPath   string
	LocalPath  string
}

type Site struct {
	Client
	Name      string
	Dir       string
	MaxAge    time.Duration
	MaxAge_   string   `json:"MaxAge"`
	Patterns_ []string `json:"Patterns"`
	Patterns  []*regexp.Regexp
}

func (s *Site) lftpCmd(cmd string) *exec.Cmd {
	args := []string{"-e", cmd + " && exit"}
	return exec.Command(s.LftpPath, args...)
}

func (s *Site) ListCmd() *exec.Cmd {
	cmd := fmt.Sprintf("cls --date --time-style='%%F %%T %%z %%Z' %s",
		s.Dir)
	return s.lftpCmd(cmd)
}

func (s *Site) GetDirs() ([]Dir, error) {
	cmd := s.ListCmd()
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
	return dirs, nil
}

func (s *Site) FilterDirs() ([]Dir, error) {
	dirs, err := s.GetDirs()
	if err != nil {
		return nil, err
	}
	res := []Dir{}
	for _, dir := range dirs {
		if dir.IsSymlink {
			continue
		}
		if !dir.CreatedAfter(s.MaxAge) {
			continue
		}
		if !dir.MatchAny(s.Patterns) {
			continue
		}
		res = append(res, dir)
	}
	return res, nil
}

func (s *Site) LocalPath(dir Dir) (string, error) {
	series, err := ParseSeries(dir.Base())
	if err != nil {
		return "", err
	}
	localPath := filepath.Join(s.Client.LocalPath, series.Name,
		"S"+series.Season)
	if !strings.HasSuffix(localPath, string(os.PathSeparator)) {
		localPath += string(os.PathSeparator)
	}
	return localPath, nil
}

func (s *Site) getCmd(dir Dir) (string, error) {
	localPath, err := s.LocalPath(dir)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", s.LftpGetCmd, dir.Path, localPath), nil
}

func (s *Site) GetCmd(dir Dir) (*exec.Cmd, error) {
	getCmd, err := s.getCmd(dir)
	if err != nil {
		return nil, err
	}
	return s.lftpCmd(getCmd), nil
}

func (s *Site) queueCmd(dirs []Dir) (string, error) {
	queueCmds := []string{}
	for _, d := range dirs {
		getCmd, err := s.getCmd(d)
		if err != nil {
			return "", err
		}
		queueCmds = append(queueCmds, "queue "+getCmd)
	}
	queueCmds = append(queueCmds, "queue start")
	queueCmds = append(queueCmds, "wait")
	return strings.Join(queueCmds, " && "), nil
}

func (s *Site) QueueCmd(dirs []Dir) (*exec.Cmd, error) {
	queueCmd, err := s.queueCmd(dirs)
	if err != nil {
		return nil, err
	}
	return s.lftpCmd(queueCmd), nil
}
