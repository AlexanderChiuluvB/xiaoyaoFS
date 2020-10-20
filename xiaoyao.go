package main

import (
	"flag"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/command"
	"os"
)

var exitStatus = 0

func main() {

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You forget to input argument!")
	}

	for _, cmd := range command.Commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			if !cmd.Run(cmd, args) {
				fmt.Fprintf(os.Stderr, "\n")
				cmd.Flag.Usage()
				fmt.Fprintf(os.Stderr, "Default Parameters:\n")
				cmd.Flag.PrintDefaults()
			}
			exit()
			return
		}
	}
}

var atexitFuncs []func()

func exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}



