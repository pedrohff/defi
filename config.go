package main

import (
	"errors"
	"flag"
	"fmt"
	"time"
)

const usageMessage = "usage: defi [--interval N] [--once] [path|pattern]"

type appConfig struct {
	spec     watchSpec
	interval time.Duration
	once     bool
}

func parseAppConfig(args []string) (appConfig, string, error) {
	fs := flag.NewFlagSet("defi", flag.ContinueOnError)
	intervalFlag := fs.Int("interval", 1, "Polling interval in seconds")
	onceFlag := fs.Bool("once", false, "Run tests once and exit")

	if err := fs.Parse(args); err != nil {
		return appConfig{}, "", err
	}

	if *intervalFlag <= 0 {
		return appConfig{}, "", fmt.Errorf("interval must be greater than zero")
	}

	remaining := fs.Args()
	target := "."
	if len(remaining) > 0 {
		target = remaining[0]
	}
	spec, err := parseWatchSpec(target)
	if err != nil {
		return appConfig{}, "", err
	}

	cfg := appConfig{
		spec:     spec,
		interval: time.Duration(*intervalFlag) * time.Second,
		once:     *onceFlag,
	}

	initialPath := ""
	if path, _, err := resolveLatestTarget(spec); err == nil {
		initialPath = path
	} else if errors.Is(err, errNoMatchingFiles) {
		if cfg.once {
			return cfg, "", fmt.Errorf("no matching files found for %s", spec.DisplayBase())
		}
	} else {
		return appConfig{}, "", err
	}

	return cfg, initialPath, nil
}
