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

func cleanWorkingTree() {
	exec.Command("git", "reset", "--hard").Run()
}

func dirtyWorkingTree() {
	exec.Command("echo", "'Dirty that working tree! Work it!'", ">>", "README.md").Run()
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
// func TestCPConfigs(t *testing.T) {
// 	dirs := getPullDirs()
// }

func TestCPGhosttyConfig(t *testing.T) {
	// I'm going to copy from src to destination but with a different name so
	// the copied file can be removed within the test and I don't have to
	// modify my ghostty config file
	fakeGhosttyFile := "config_test"
	ghosttyDestination := getGhosttyDestination()
	ghosttySrc := getPullDirs()[0] + "ghostty/"

	chdir(ghosttySrc)
	exec.Command("touch", fakeGhosttyFile).Run()

	_, cpError := cpGhosttyConfig(ghosttySrc + fakeGhosttyFile)
	if cpError != nil {
		t.Errorf("There was a `cp` error [%v]", cpError)
		t.FailNow()
	}

	chdir(ghosttyDestination)
	_, stderr := exec.Command("ls", fakeGhosttyFile).Output()
	if stderr != nil {
		t.Errorf("There was an error running `ls` [%v]", stderr)
	}

	chdir(ghosttySrc)
	exec.Command("rm", fakeGhosttyFile).Run()

	chdir(ghosttyDestination)
	exec.Command("rm", fakeGhosttyFile).Run()
}

func TestCPVimConfig(t *testing.T) {
	// I'm going to copy from src to destination but with a different name so
	// the copied file can be removed within the test and I don't have to
	// modify my vim config file
	fakeVimFile := ".vimrc_test"
	vimDestination := getHomePath()
	vimSrc := getPullDirs()[0] + "vim/"

	chdir(vimSrc)
	exec.Command("touch", fakeVimFile).Run()

	_, cpError := cpVimConfig(vimSrc + fakeVimFile)
	if cpError != nil {
		t.Errorf("There was a `cp` error [%v]", cpError)
		t.FailNow()
	}

	chdir(vimDestination)
	_, stderr := exec.Command("ls", fakeVimFile).Output()
	if stderr != nil {
		t.Errorf("There was an error running `ls` [%v]", stderr)
	}

	chdir(vimSrc)
	exec.Command("rm", fakeVimFile).Run()

	chdir(vimDestination)
	exec.Command("rm", fakeVimFile).Run()
}

func TestGetConfigs(t *testing.T) {
	dirs := getPullDirs()

	// Happy path

	for _, dir := range dirs {
		chdir(dir)

		exec.Command("git", "stash").Run()
	}

	statusErrors, pullErrors := getConfigs(dirs)

	if len(statusErrors) != 0 {
		t.Error("There were git status errors", statusErrors)
	}

	if len(pullErrors) != 0 {
		t.Error("There were git pull errors", pullErrors)
	}

	for _, dir := range dirs {
		chdir(dir)

		exec.Command("git", "stash", "apply").Run()
		exec.Command("git", "stash", "clear").Run()
	}

	// Sad path

	for _, dir := range dirs {
		chdir(dir)

		// NOTE: if there aren't changes upstream, "Already up to date" is returned
		// regardless of what we stash/change
		exec.Command("git", "fetch").Run()
		stdout, stderr := exec.Command("git", "status").Output()
		if stderr != nil {
			t.Error("Unexpected error checking git status", stderr)
		}

		if strings.Contains(string(stdout), "Your branch is behind 'origin/") {
			exec.Command("git", "stash").Run()
			exec.Command("echo", "'New Change'", ">>", "test.txt").Run()

			statusErrors, pullErrors = getConfigs(dirs)

			if len(statusErrors) == 0 {
				t.Error("There were git status errors", statusErrors)
			}

			if len(pullErrors) == 0 {
				t.Error("There were git pull errors", pullErrors)
			}

			for _, dir := range dirs {
				chdir(dir)

				exec.Command("git", "reset", "--hard").Run()
				exec.Command("git", "stash", "apply").Run()
				exec.Command("git", "stash", "clear").Run()
			}
		} else {
			// TODO: how do we handle sad path when there are no upstream changes???
		}
	}
}

func TestGetGitStatus(t *testing.T) {
	happyPath := "nothing to commit, working tree clean"
	path := getHomePath() + "/dev/configpp"

	// Happy path

	chdir(path)

	gitStashBegin()

	stdout, stderr := getGitStatus(path)
	if stderr != nil {
		t.Error("There was an error calling getGitStatus()", stderr)
	}

	stdoutString := string(stdout)
	if !strings.Contains(stdoutString, happyPath) {
		t.Errorf("Expected getGitStatus() to contain [%s], but it returned [%s]", happyPath, stdoutString)
	}

	gitStashEnd()

	// Sad path

	gitStashBegin()
	dirtyWorkingTree()

	stdout, stderr = getGitStatus(path)
	if stderr != nil {
		t.Error("Unexpected getGitStatus() to throw an error")
	}

	if strings.Contains(string(stdout), "untracked files") {
		t.Error("Expected getGetStatus() to return untracked files", stdout)
	}

	cleanWorkingTree()
	gitStashEnd()
}

func TestGetPullDirs(t *testing.T) {
	dirs := getPullDirs()

	if dirs[0] != getHomePath()+"/dev/configs/" {
		t.Error("getPullDirs() returned the wrong path for configs")
	}

	if dirs[1] != getHomePath()+"/.config/nvim/" {
		t.Error("getPullDirs() returned the wrong path for nvim")
	}
}

func TestPullFromGit(t *testing.T) {
	path := getHomePath() + "/dev/configpp"

	// Happy path - clear all changes from git status

	gitStashBegin()

	_, stderr := pullFromGit(path)
	if stderr != nil {
		t.Errorf("There was an unexpected error from pullFromGit(): [%s]", stderr.Error())
	}

	gitStashEnd()

	// Sad path - create uncommitted changes to dirty the working tree to prevent pulling

	gitStashBegin()
	dirtyWorkingTree()

	_, stderr = pullFromGit(path)
	if stderr == nil {
		t.Errorf("Error expected from pullFromGit()")
	}

	cleanWorkingTree()
	gitStashEnd()
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
