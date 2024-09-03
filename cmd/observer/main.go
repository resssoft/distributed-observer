package main

import (
	"context"
	"flag"
	"os"

	"observer/internal/manager"
	"observer/internal/settings"
)

func main() {
	showVer := flag.Bool("v", false, "show version")
	debugMode := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	if *showVer {
		print(settings.Version())
		os.Exit(0)
	}

	m := manager.New()
	if *debugMode {
		println(settings.Version())
	}
	m.Start(context.Background())
	println("exit")
}
