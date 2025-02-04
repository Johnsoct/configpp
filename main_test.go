package main

import (
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
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
func TestGetGitStatus(t *testing.T) {
	happyPath := "nothing to commit, working tree clean"
	path := getHomePath() + "/dev/configpp"
	sadPath := "Changes not staged for commit"

	// Happy path

	stdout, stderr := getGitStatus(path)
	if stderr != nil {
		t.Error("There was an error calling getGitStatus()", stderr)
	}

	stdoutString := string(stdout)
	if stdoutString != happyPath {
		t.Error("The output of getGitStatus was not as expected", stdoutString)
	}

	// Sad path

	exec.Command("touch", "test.txt").Output()

	stdout, stderr = getGitStatus(path)
	if stderr != nil {
		t.Error("There was an error calling getGitStatus()", stderr)
	}

	stdoutString = string(stdout)
	if !strings.Contains(stdoutString, sadPath) {
		t.Error("The output of getGitStatus() was not as expected", stdoutString, sadPath)
	}
}

func TestGetPullDirs(t *testing.T) {
	dirs := getPullDirs()

	if dirs[0] != getHomePath()+"dev/configs/eslint/" {
		t.Error("getPullDirs() returned the wrong path for eslint")
	}

	if dirs[1] != getHomePath()+"dev/configs/ghostty/" {
		t.Error("getPullDirs() returned the wrong path for ghostty")
	}

	if dirs[2] != getHomePath()+".config/nvim/" {
		t.Error("getPullDirs() returned the wrong path for nvim")
	}

	if dirs[3] != getHomePath()+"dev/configs/stylelint/" {
		t.Error("getPullDirs() returned the wrong path for stylelint")
	}

	if dirs[4] != getHomePath()+"dev/configs/vim/" {
		t.Error("getPullDirs() returned the wrong path for vim")
	}
}

func TestPullFromGit(t *testing.T) {
	tests := []InputOutput{
		{input: getHomePath() + "/dev/configpp", output: "Already up to date."},
	}

	for _, v := range tests {
		_, stderr := pullFromGit(v.input)

		if stderr != nil {
			t.Error("There was an unexpected error from pullFromGit()")
		}
	}
}

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
