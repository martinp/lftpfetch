package main

import (
	"fmt"
	"io"
	"os"

	flags "github.com/jessevdk/go-flags"

	"github.com/mpolden/lftpq/lftp"
	"github.com/mpolden/lftpq/queue"
)

type lister interface {
	List(name, path string) ([]os.FileInfo, error)
}

type CLI struct {
	Config   string `short:"f" long:"config" description:"Path to config" value-name:"FILE" default:"~/.lftpqrc"`
	Dryrun   bool   `short:"n" long:"dryrun" description:"Print queue and exit"`
	Format   string `short:"F" long:"format" description:"Format to use in dryrun mode" choice:"lftp" choice:"json" default:"lftp"`
	Test     bool   `short:"t" long:"test" description:"Test and print config"`
	Quiet    bool   `short:"q" long:"quiet" description:"Do not print output from lftp"`
	Import   bool   `short:"i" long:"import" description:"Build queues from stdin"`
	LftpPath string `short:"p" long:"lftp" description:"Path to lftp program" value-name:"NAME" default:"lftp"`
	consumer queue.Consumer
	lister   lister
	wr       io.Writer
	rd       io.Reader
}

func (c *CLI) Run() error {
	cfg, err := queue.ReadConfig(c.Config)
	if err != nil {
		return err
	}
	if c.Test {
		json, err := cfg.JSON()
		if err != nil {
			return err
		}
		fmt.Fprintf(c.wr, "%s\n", json)
		return nil
	}
	var queues []queue.Queue
	if c.Import {
		queues, err = queue.Read(cfg.Sites, c.rd)

	} else {
		queues, err = c.queueFor(cfg.Sites)
	}
	if err != nil {
		return err
	}
	for _, q := range queues {
		if err := c.transfer(q); err != nil {
			c.printf("error while processing queue for %s: %s\n", q.Site.Name, err)
			continue
		}
	}
	return nil
}

func (c *CLI) printf(format string, vs ...interface{}) {
	alwaysPrint := false
	for _, v := range vs {
		if _, ok := v.(error); ok {
			alwaysPrint = true
			break
		}
	}
	if !c.Quiet || alwaysPrint {
		fmt.Fprint(c.wr, "lftpq: ")
		fmt.Fprintf(c.wr, format, vs...)
	}
}

func (c *CLI) queueFor(sites []queue.Site) ([]queue.Queue, error) {
	var queues []queue.Queue
	for _, s := range sites {
		if s.Skip {
			c.printf("skipping site %s\n", s.Name)
			continue
		}
		var files []os.FileInfo
		for _, dir := range s.Dirs {
			f, err := c.lister.List(s.Name, dir)
			if err != nil {
				c.printf("error while listing %s on %s: %s\n", dir, s.Name, err)
				continue
			}
			files = append(files, f...)
		}
		queue := queue.New(s, files)
		queues = append(queues, queue)
	}
	return queues, nil
}

func (c *CLI) transfer(q queue.Queue) error {
	if c.Dryrun {
		var (
			out []byte
			err error
		)
		if c.Format == "json" {
			out, err = q.MarshalJSON()
			out = append(out, 0x0a) // Add trailing newline
		} else {
			out, err = q.MarshalText()
		}
		if err == nil {
			fmt.Fprintf(c.wr, "%s", out)
		}
		return err
	}
	if len(q.Transferable()) == 0 {
		c.printf("%s queue is empty\n", q.Site.Name)
		return nil
	}
	if err := q.Start(c.consumer); err != nil {
		return err
	}
	if q.Site.PostCommand == "" {
		return nil
	}
	cmd, err := q.PostCommand(!c.Quiet)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func main() {
	var cli CLI
	_, err := flags.ParseArgs(&cli, os.Args)
	if err != nil {
		os.Exit(1)
	}
	cli.wr = os.Stdout
	cli.rd = os.Stdin
	client := lftp.Client{Path: cli.LftpPath, InheritIO: !cli.Quiet}
	cli.lister = &client
	cli.consumer = &client
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
