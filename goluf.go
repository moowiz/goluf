package main

// goluf checks if a go package has been modified before a certain time.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Expected 3 args but got %v\n", os.Args)
		os.Exit(1)
	}

	lastMod, err := getTransitiveLastUpdated(os.Args[1])
	if err != nil {
		panic(err)
	}
	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	cmpTime := time.Unix(int64(num), 0)
	if lastMod.After(cmpTime) {
		os.Exit(1)
	}
	os.Exit(0)
}

// getTransitiveLastUpdated gets the most recent time any dependency of a given golang module has been
// updated.
func getTransitiveLastUpdated(path string) (time.Time, error) {
	args := append([]string{"list", "-f", "{{.Standard}}|{{.Dir}}|{{.GoFiles}}"}, "-deps", path)
	cmd := exec.Command("go", args...)
	out, err := cmd.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if ok {
			// FIXME: Change to logging
			fmt.Printf("stderr: \"%v\"\n", string(ee.Stderr))
			fmt.Printf("stdout: %v\n", string(out))
		}
		return time.Time{}, err
	}

	allFiles := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		sp := strings.Split(line, "|")

		// Standard library, we don't care.
		if sp[0] == "true" {
			continue
		}

		dir := sp[1]
		files := strings.Split(strings.Trim(sp[2], "[]"), " ")
		for _, f := range files {
			allFiles = append(allFiles, filepath.Join(dir, f))
		}
	}

	latestTime := time.Time{}
	for _, f := range allFiles {
		t := getFileTime(f)
		if latestTime.Before(t) {
			latestTime = t
		}
	}
	return latestTime, nil
}

func getFileTime(path string) time.Time {
	data, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return data.ModTime()
}
