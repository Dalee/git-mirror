package main

import (
	"fmt"
	"os"
)

// start here
func main() {
	m, err := parseCommandLine()
	if err != nil {
		m.displayUsage()
	}

	// load, parse and validate repository list
	if err = m.loadRepositoryList(); err != nil {
		panic(err)
	}

	// print summary report
	fmt.Println("> Found repositories:", len(m.repoList))
	fmt.Println("> Concurrency:", m.concurrentJobs)
	if m.cleanupCacheDir {
		fmt.Println("> Cache is disabled, using temporary directory...")
	} else {
		fmt.Println("> Cache will be stored in:", m.cacheDir)
	}

	// perform sync
	ok := true
	chDone, chOut, chErr := m.process()
	for {
		select {
		case line := <-chOut:
			fmt.Println(line)

		case err := <-chErr:
			fmt.Println(err.Error())
			ok = false

		case <-chDone:
			if ok {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
	}
}
