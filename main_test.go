package main

import (
	"os"
	"os/user"
	"runtime"
	"testing"
)

var (
	USER, _  = user.Current()
	USERNAME = USER.Username
)

type InputOutput struct {
	input  string
	output string
}

func TestChdir(t *testing.T) {
	home := getHomePath()
	tests := []InputOutput{
		{input: home + "/dev", output: home + "/dev"},
		{input: "~/dev", output: home + "/dev"},
		{input: "~/dev/configs/vim", output: home + "/dev/configs/vim"},
	}

	for _, test := range tests {
		chdir(test.input)

		cwd, err := os.Getwd()
		if err != nil {
			t.Errorf("There was an error with chcwd; test: %s; err: %v", test, err)
		}

		if cwd != test.output {
			t.Errorf("chdir did not change to the expected cwd; cwd: %s; expect: %s", cwd, test)
		}
	}
}

// TODO:
func TestCPConfigs(t *testing.T) {}

// TODO:
func TestCPGhosttyConfig(t *testing.T) {}

// TODO:
func TestCPVimConfig(t *testing.T) {}

// TODO:
func TestGetConfigs(t *testing.T) {}

// TODO:
func TestGetGitStatus(t *testing.T) {}

// TODO:
func TestGetPullDirs(t *testing.T) {}

// TODO:
func TestPullFromGit(t *testing.T) {}

func TestGetHomePath(t *testing.T) {
	OS := runtime.GOOS

	if OS == "linux" {
		expect := "/home/" + USERNAME
		path := getHomePath()
		if path != expect {
			t.Errorf("Homepath returned for Linux is incorrect")
		}
	} else if OS == "darwin" {
		expect := "/Users/" + USERNAME
		path := getHomePath()
		if path != expect {
			t.Errorf("Homepath returned for Mac is incorrect")
		}
	}
}

func TestReplaceTildeInPath(t *testing.T) {
	home := getHomePath()
	tests := []InputOutput{
		{input: "~/dev", output: home + "/dev"},
		{input: "~/dev/configs/vim", output: home + "/dev/configs/vim"},
	}

	for _, v := range tests {
		replaced, err := replaceTildeInPath(v.input)
		if err != nil {
			t.Error("There was an issue internally of replaceTildeInPath()")
		}

		if replaced != v.output {
			t.Errorf("Tilde in path was not replaced properly; test: %s; replaced: %s", v.input, replaced)
		}
	}
}
