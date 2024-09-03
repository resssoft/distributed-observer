package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"

	"observer/pkg/version"
)

func main() {
	var opts struct {
		ConfigFile  string `short:"c" long:"config" description:"Config filepath" default:"config.yml"`
		ConfigWrite bool   `short:"w" long:"writeconfig" description:"Write config to filepath"`
		UseReflect  bool   `long:"reflect" description:"Use gRPC reflection"`
		ShowVersion func() `short:"v" long:"version" description:"Show version and exit" json:"-"`
	}
	opts.ShowVersion = showVersion
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatalf("flags: %w", err)
	}
}

func showVersion() {
	// nolint: forbidigo
	fmt.Printf("Version: reciver-%s\n", version.Get())
	os.Exit(0)
}
