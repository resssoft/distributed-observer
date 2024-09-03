package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/minio/selfupdate"

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
	fmt.Printf("Version: sender-%s\n", version.Get())
	os.Exit(0)
}

func doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// error handling
	}
	return err
}
