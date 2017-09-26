package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"sync"
	"crypto/sha1"
	"os/exec"
	"errors"
	"path"
	"time"
	"strings"
)

// checkout helper
func getCheckoutDir(workDir string, src string) string {
	return fmt.Sprintf("%s/%x", workDir, sha1.Sum([]byte(src)))
}

// exec helper
func executeGit(commandDir string, cmdArgs ...string) error {
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = commandDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(output))
	}

	return nil
}

// the base sync scenario
func syncRepo(cfg Config, src string, dst string) error {
	cloneDir := getCheckoutDir(cfg.CacheDir, src)
	headPath := path.Join(cloneDir, "HEAD")

	_, statResult := os.Stat(headPath)
	if os.IsNotExist(statResult) {
		if err := executeGit("", "clone", "--mirror", src, cloneDir); err != nil {
			return err
		}
	} else {
		if err := executeGit(cloneDir, "fetch", "-p", "origin"); err != nil {
			return err
		}
	}

	mirrorScenario := [][]string{
		{"symbolic-ref", "HEAD", "refs/heads/master"},
		{"remote", "set-url", "origin", "--push", dst},
		{"push", "--mirror"},
	}

	for _, args := range mirrorScenario {
		if err := executeGit(cloneDir, args...); err != nil {
			return err
		}
	}

	if cfg.CleanupCacheDir {
		return os.RemoveAll(cloneDir)
	}

	return nil
}

// error exit helper
func errorExit(errorList []string) {
	fmt.Println("> Errors detected:")
	fmt.Println(strings.Join(errorList, "\n"))
	os.Exit(1)
}

// start here
func main() {
	var repoList []map[string]interface{}
	var waitList sync.WaitGroup

	cfg, err := GetConfig()
	if err != nil {
		if HasWrongArguments(err) {
			cfg.DisplayHelp()
			os.Exit(0)
		} else {
			panic(err)
		}
	}

	jsonBytes, err := ioutil.ReadFile(cfg.JsonListFile)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonBytes, &repoList); err != nil {
		panic(err)
	}

	if errorList, ok := cfg.IsValidRepoList(repoList); !ok {
		errorExit(errorList)
	}

	fmt.Println("> Found repositories:", len(repoList))
	fmt.Println("> Concurrency:", cfg.ConcurrentJobs)
	if cfg.CleanupCacheDir {
		fmt.Println("> Cache is disabled, using temporary directory...")
	} else {
		fmt.Println("> Cache will be stored in:", cfg.CacheDir)
	}

	guardChan := make(chan bool, cfg.ConcurrentJobs)
	errorList := make([]string, 0)
	startedAt := time.Now()

	for _, item := range repoList {
		waitList.Add(1)
		src, _ := item[cfg.SrcKey].(string)
		dst, _ := item[cfg.DstKey].(string)

		go func(src string, dst string) {
			defer func() {
				waitList.Done()
				<-guardChan
			}()

			guardChan <- true

			// mirror repository!
			started := time.Now()
			err := syncRepo(cfg, src, dst)
			elapsed := time.Now().Sub(started).Round(time.Second)
			if err != nil {
				errorList = append(errorList, err.Error())
			}
			fmt.Println("+", src, "in", elapsed)

		}(src, dst)
	}

	// wait for goroutines...
	waitList.Wait()

	// print final report
	fmt.Println("> Finished in", time.Now().Sub(startedAt).Round(time.Second))
	if len(errorList) > 0 {
		errorExit(errorList)
	}
}
