package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/martinp/lftptv/site"
	"log"
	"os"
)

func main() {
	var opts struct {
		Dryrun bool   `short:"n" long:"dryrun" description:"Print generated command instead of running it"`
		Config string `short:"f" long:"config" description:"Path to config" value-name:"FILE" required:"true"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}
	cfg, err := site.ReadConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range cfg.Sites {
		dirs, err := s.FilterDirs()
		if err != nil {
			log.Fatal(err)
		}
		cmds := []string{}
		for _, d := range dirs {
			cmd, err := s.QueueCmd(d)
			if err != nil {
				log.Printf("Failed to create queue cmd for %s: %s", d.Path, err)
				continue
			}
			cmds = append(cmds, cmd)
		}
		fmt.Println(s.QueueCmds(cmds))
	}
}
