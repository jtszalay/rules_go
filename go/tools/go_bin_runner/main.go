package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

var GoBinRlocationPath = "not set"
var GoEnvJsonRlocationPath = "not set"

func main() {
	goBin, err := runfiles.Rlocation(GoBinRlocationPath)
	if err != nil {
		log.Fatal(err)
	}
	// The go binary lies at $GOROOT/bin/go.
	goRoot, err := filepath.Abs(filepath.Dir(filepath.Dir(goBin)))
	if err != nil {
		log.Fatal(err)
	}

	env, err := getUserGoEnv()
	if err != nil {
		log.Fatal(err)
	}
	// Override GOROOT to point to the hermetic Go SDK.
	env = append(env, "GOROOT="+goRoot)

	args := append([]string{goBin}, os.Args[1:]...)
	log.Fatal(runProcess(args, env, os.Getenv("BUILD_WORKING_DIRECTORY")))
}

func getUserGoEnv() ([]string, error) {
	var goEnv map[string]string
	// Special value set when rules_go is loaded as a WORKSPACE repo, in which
	// the user-configured Go env isn't available.
	if GoEnvJsonRlocationPath != "WORKSPACE" {
		goEnvJsonPath, err := runfiles.Rlocation(GoEnvJsonRlocationPath)
		if err != nil {
			return nil, err
		}
		goEnvJson, err := os.ReadFile(goEnvJsonPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(goEnvJson, &goEnv)
		if err != nil {
			return nil, err
		}
	}

	if len(goEnv) == 0 {
		// Fall back to the environment of the current process if there is no
		// use of the go_sdk.env tag. This was the default before the tag was
		// introduced.
		return os.Environ(), nil
	} else {
		var env []string
		for k, v := range goEnv {
			env = append(env, k+"="+v)
		}
		return env, nil
	}
}

func runProcess(args, env []string, dir string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		os.Exit(exitErr.ExitCode())
	} else if err == nil {
		os.Exit(0)
	}
	return err
}
