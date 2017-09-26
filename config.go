package main

import (
	"flag"
	"os"
	"errors"
	"fmt"
)

type (
	Config struct {
		CacheDir        string
		CleanupCacheDir bool
		ConcurrentJobs  int
		JsonListFile    string
		SrcKey          string
		DstKey          string
	}
)

var (
	errWrongArguments  = errors.New("wrong number of arguments")
	txtCacheDir        = "Cache directory"
	txtCleanupCacheDir = "Cache cleanup (automatic when cache directory is not provided)"
	txtConcurrency     = "Maximum concurrent workers"
)

// is help or mistyped?
func HasWrongArguments(err error) bool {
	return err == errWrongArguments
}

// create config from command line
func GetConfig() (Config, error) {
	var helpRequested bool

	cfg := Config{}
	flag.StringVar(&cfg.CacheDir, "cacheDir", "", txtCacheDir)
	flag.BoolVar(&cfg.CleanupCacheDir, "cleanCache", false, txtCleanupCacheDir)
	flag.IntVar(&cfg.ConcurrentJobs, "concurrency", 5, txtConcurrency)
	flag.BoolVar(&helpRequested, "help", false, "This help")
	flag.Parse()

	// if help requested, just return error and empty cfg
	if helpRequested {
		return cfg, errWrongArguments
	}

	if cfg.CacheDir == "" {
		cfg.CleanupCacheDir = true
		cfg.CacheDir = os.TempDir()
	}

	if flag.NArg() != 3 {
		return cfg, errWrongArguments
	}

	cfg.JsonListFile = flag.Arg(0)
	cfg.SrcKey = flag.Arg(1)
	cfg.DstKey = flag.Arg(2)
	return cfg, nil
}

// display help
func (c *Config) DisplayHelp() {
	fmt.Println("Usage: git-mirror [OPTIONS] list.json srcKey dstKey")
	flag.PrintDefaults()
}

// validate json
func (c *Config) IsValidRepoList(repoList []map[string]interface{}) ([]string, bool) {
	errorList := make([]string, 0)

	for _, item := range repoList {
		if _, ok := item[c.SrcKey].(string); !ok {
			errorList = append(
				errorList,
				fmt.Sprintf("- key '%s' not found or not a string", c.SrcKey),
			)
		}
		if _, ok := item[c.DstKey].(string); !ok {
			errorList = append(
				errorList,
				fmt.Sprintf("- key '%s' not found or not a string", c.DstKey),
			)
		}
	}

	return errorList, len(errorList) == 0
}
