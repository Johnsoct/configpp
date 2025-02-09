package main

import (
	"os"
	"os/exec"
	"os/user"
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
func TestCPConfig(t *testing.T) {
	baseDest := "/tmp"
	baseSrc := replaceTildeInPath("~/dev/configpp")
	fauxFile := "dummy.txt"
	dest := baseDest + "/" + fauxFile
	src := baseSrc + "/" + fauxFile

	happyPath := Config{
		destLinux: dest,
		destMac:   dest,
		src:       src,
	}
	sadSrcPath := Config{
		destLinux: dest,
		destMac:   dest,
		src:       "~/dev/configpp/does-not-exist.txt",
	}
	sadDestPath := Config{
		destLinux: dest + "badddddddd",
		destMac:   dest + "badddddddd",
		src:       src,
	}

	// Happy path (toDest)

	exec.Command("touch", fauxFile).Run()

	stdout, stderr := cpConfig(happyPath, true)
	if stderr != nil {
		t.Errorf("There was an unexpected error copying:\n%s\n", stdout)
	}

	stdout, stderr = exec.Command("ls", baseDest).CombinedOutput()
	if stderr != nil {
		t.Errorf("There was an unexpected error checking the copy test result:\n%s\n", stdout)
	}

	if !strings.Contains(string(stdout), fauxFile) {
		t.Errorf("The faux file was not found in the destination")
	}

	exec.Command("rm", dest).Run()
	exec.Command("rm", src).Run()

	// Happy path (toSrc)

	exec.Command("touch", dest).Run()

	stdout, stderr = cpConfig(happyPath, false)
	if stderr != nil {
		t.Errorf("There was an unexpected error copying:\n%s\n", stdout)
	}

	stdout, stderr = exec.Command("ls", baseSrc).CombinedOutput()
	if stderr != nil {
		t.Errorf("There was an unexpected error checking the copy test result:\n%s\n", stdout)
	}

	if !strings.Contains(string(stdout), fauxFile) {
		t.Errorf("The faux file was not found in the src")
	}

	exec.Command("rm", dest).Run()
	exec.Command("rm", src).Run()

	// Sad path (src filepath)

	exec.Command("touch", src).Run()

	_, stderr = cpConfig(sadSrcPath, true)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad src filepath")
	}

	exec.Command("rm", src).Run()

	// Sad path (dest filepath)

	exec.Command("touch", dest).Run()

	_, stderr = cpConfig(sadDestPath, true)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad dest filepath")
	}

	exec.Command("rm", dest).Run()
}

func TestGetConfigs(t *testing.T) {
	// Happy path

	for _, dir := range ConfigsSrc {
		chdir(dir)
		gitStashBegin()
	}

	statusErrors, pullErrors := getConfigs()

	if len(statusErrors) != 0 {
		t.Error("There were git status errors", statusErrors)
	}

	if len(pullErrors) != 0 {
		t.Error("There were git pull errors", pullErrors)
	}

	for _, dir := range ConfigsSrc {
		chdir(dir)
		gitStashEnd()
	}

	// Sad path

	for _, dir := range ConfigsSrc {
		chdir(dir)

		// NOTE: if there aren't changes upstream, "Already up to date" is returned
		// regardless of what we stash/change
		exec.Command("git", "fetch").Run()
		stdout, stderr := exec.Command("git", "status").Output()
		if stderr != nil {
			t.Error("Unexpected error checking git status", stderr)
		}

		if strings.Contains(string(stdout), "Your branch is behind 'origin/") {
			gitStashBegin()
			dirtyWorkingTree()

			statusErrors, pullErrors = getConfigs()

			if len(statusErrors) == 0 {
				t.Error("There were git status errors", statusErrors)
			}

			if len(pullErrors) == 0 {
				t.Error("There were git pull errors", pullErrors)
			}

			for _, dir := range ConfigsSrc {
				chdir(dir)
				cleanWorkingTree()
				gitStashEnd()
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

	_, stderr = getGitStatus(path)
	if stderr != nil {
		t.Error("Unexpected getGitStatus() to throw an error")
	}

	cleanWorkingTree()
	gitStashEnd()
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
	if stderr != nil {
		t.Errorf("Error expected from pullFromGit()")
	}

	cleanWorkingTree()
	gitStashEnd()
}

func TestGetHomePath(t *testing.T) {
	if OS == "darwin" {
		expect := "/Users/" + USERNAME
		if getHomePath() != expect {
			t.Errorf("Homepath returned for Mac is incorrect")
		}
	} else {
		expect := "/home/" + USERNAME
		if getHomePath() != expect {
			t.Errorf("Homepath returned for Linux is incorrect")
		}
	}
}

func TestReplaceTildeInPath(t *testing.T) {
	home := getHomePath()
	tests := []InputOutput{
		{input: "~/dev", output: home + "/dev"},
		{input: "~/dev/configs/vim", output: Vim.src},
	}

	for _, v := range tests {
		replaced := replaceTildeInPath(v.input)

		if replaced != v.output {
			t.Errorf("Tilde in path was not replaced properly; test: %s; replaced: %s", v.input, replaced)
		}
	}
}
