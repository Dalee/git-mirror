package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type (
	mirror struct {
		cacheDir        string
		cleanupCacheDir bool
		concurrentJobs  int
		jsonListFile    string
		srcKey          string
		dstKey          string
		refs            refs

		repoList []repo
	}

	repo struct {
		dir string
		src string
		dst string
	}

	refs []string
)

func (r *refs) String() string {
	return "refs"
}

func (r *refs) Set(value string) error {
	*r = append(*r, value)
	return nil
}

// git exec helper
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(output))
	}

	return nil
}

// perform repository mirroring
func (r *repo) sync(cacheDir string, cleanupOnSuccess bool, refs *refs) error {
	var scenario [][]string

	// check if repository exists
	_, s := os.Stat(path.Join(r.dir, "HEAD"))
	if os.IsNotExist(s) {
		if err := runGitCommand(cacheDir, "clone", "--mirror", r.src, r.dir); err != nil {
			return err
		}
	} else {
		scenario = append(scenario, []string{r.dir, "fetch", "-p", "origin"})
	}

	if len(*refs) > 0 {
		runGitCommand(r.dir, "config", "--unset-all", "remote.origin.push")

		for _, ref := range *refs {
			refSpec := fmt.Sprintf("+refs/%s/*:refs/%s/*", ref, ref)
			scenario = append(scenario, []string{r.dir, "config", "--add", "remote.origin.push", refSpec})
		}
	}

	// add rest of commands..
	scenario = append(scenario, [][]string{
		{r.dir, "symbolic-ref", "HEAD", "refs/heads/master"},
		{r.dir, "remote", "set-url", "origin", "--push", r.dst},
		{r.dir, "push", "--mirror"},
	}...)

	for _, args := range scenario {
		if err := runGitCommand(args[0], args[1:]...); err != nil {
			return err
		}
	}

	if cleanupOnSuccess {
		return os.RemoveAll(r.dir)
	}

	return nil
}

// parse and validate provided command line arguments
func parseCommandLine() (mirror, error) {
	var helpRequested bool

	cfg := mirror{}
	flag.Usage = func() {
		fmt.Printf("Usage: %s ", os.Args[0])
		fmt.Printf("[OPTIONS] repository_list.json srcKey dstKey\n\n")
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&cfg.cacheDir, "cacheDir", "", "Cache directory")
	flag.BoolVar(&cfg.cleanupCacheDir, "cleanCache", false, "Cache cleanup (automatic when cache directory is not provided)")
	flag.IntVar(&cfg.concurrentJobs, "concurrency", 5, "Number of workers")
	flag.BoolVar(&helpRequested, "help", false, "This help")
	flag.Var(&cfg.refs, "ref", "Refs to mirror (default all)")
	flag.Parse()

	// if help requested or argument mismatch count, just exit with usage
	if helpRequested || flag.NArg() != 3 {
		flag.Usage()
		os.Exit(0)
	}

	if cfg.cacheDir == "" {
		cfg.cleanupCacheDir = true
		cfg.cacheDir = os.TempDir()
	}

	cfg.jsonListFile = flag.Arg(0)
	cfg.srcKey = flag.Arg(1)
	cfg.dstKey = flag.Arg(2)

	return cfg, nil
}

// load and validate json
func (c *mirror) loadRepositoryList() error {
	var confRawList []map[string]interface{}

	jsonBytes, err := ioutil.ReadFile(c.jsonListFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonBytes, &confRawList); err != nil {
		return err
	}

	for _, item := range confRawList {
		src, _ := item[c.srcKey].(string)
		dst, _ := item[c.dstKey].(string)

		c.repoList = append(c.repoList, repo{
			dir: fmt.Sprintf("%s/%x", c.cacheDir, sha1.Sum([]byte(src))),
			src: src,
			dst: dst,
		})
	}

	return c.validate()
}

// validate slice of structures from parsed json
func (c *mirror) validate() error {
	errorList := make([]string, 0)

	for _, item := range c.repoList {
		if item.src == "" {
			errorList = append(
				errorList,
				fmt.Sprintf("- key '%s' not found or not a string", c.srcKey),
			)
		}
		if item.dst == "" {
			errorList = append(
				errorList,
				fmt.Sprintf("- key '%s' not found or not a string", c.dstKey),
			)
		}
	}

	if len(errorList) > 0 {
		return errors.New(strings.Join(errorList, "\n"))
	}

	return nil
}

// perform mirroring process
func (c *mirror) process() (chan bool, chan string, chan error) {
	var wg sync.WaitGroup

	chGuard := make(chan bool, c.concurrentJobs)
	chOut := make(chan string, c.concurrentJobs)
	chErr := make(chan error, c.concurrentJobs)
	chDone := make(chan bool)
	startedAt := time.Now()

	// start workers
	for _, item := range c.repoList {
		wg.Add(1)

		go func(item repo) {
			defer func() {
				<-chGuard
				wg.Done()
			}()

			chGuard <- true
			started := time.Now()
			if err := item.sync(c.cacheDir, c.cleanupCacheDir, &c.refs); err != nil {
				chErr <- err
			}

			elapsed := time.Since(started).Round(time.Second)
			chOut <- fmt.Sprintf("+ %s in %s", item.src, elapsed)

		}(item)
	}

	// wait and finalize
	go func() {
		wg.Wait()

		chOut <- fmt.Sprintf("> Finished in %s", time.Since(startedAt).Round(time.Second))
		chDone <- true

		close(chOut)
		close(chErr)
		close(chDone)
	}()

	return chDone, chOut, chErr
}
