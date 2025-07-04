package main

import (
	"os"
	"os/exec"
	"os/user"
	"path"
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

// TEST: Missing tests when the target of cp/rsync is missing the directory
// being targeted (in both directions)
// Lines 121-127 of main.go
func TestCPConfig(t *testing.T) {
	baseDest := "/tmp"
	baseSrc := replaceTildeInPath("~/dev/configpp")
	fauxFile := "dummy.txt"
	localDir := baseDest + "/" + fauxFile
	localRepo := baseSrc + "/" + fauxFile

	happyPath := Config{
		dir:       false,
		localDir:  []string{localDir, localDir},
		localRepo: localRepo,
	}
	sadSrcPath := Config{
		dir:       false,
		localDir:  []string{localDir, localDir},
		localRepo: "~/dev/configpp/does-not-exist.txt",
	}
	sadDestPath := Config{
		dir:       false,
		localDir:  []string{localDir + "badddddddd", localDir + "badddddddd"},
		localRepo: localRepo,
	}

	// Happy path (to local dir)

	exec.Command("touch", localRepo).Run()

	stdout, stderr := cpConfig(happyPath, false)
	if stderr != nil {
		t.Errorf("There was an unexpected error copying:\n%s\n", stdout)
	}

	stdout, stderr = exec.Command("ls", path.Dir(happyPath.localDir[0])).CombinedOutput()
	if stderr != nil {
		t.Errorf("There was an unexpected error checking the copy test result:\n%s\n", stdout)
	}

	if !strings.Contains(string(stdout), fauxFile) {
		t.Errorf("The faux file was not found")
	}

	exec.Command("rm", localDir).Run()
	exec.Command("rm", localRepo).Run()

	// Happy path (to local repo)

	exec.Command("touch", localDir).Run()

	stdout, stderr = cpConfig(happyPath, true)
	if stderr != nil {
		t.Errorf("There was an unexpected error copying:\n%s\n", stdout)
	}

	stdout, stderr = exec.Command("ls", baseSrc).CombinedOutput()
	if stderr != nil {
		t.Errorf("There was an unexpected error checking the copy test result:\n%s\n", stdout)
	}

	if !strings.Contains(string(stdout), fauxFile) {
		t.Errorf("The faux file was not found")
	}

	exec.Command("rm", localDir).Run()
	exec.Command("rm", localRepo).Run()

	// Sad path (to bad local repo filepath)

	exec.Command("touch", localRepo).Run()

	_, stderr = cpConfig(sadSrcPath, true)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad src filepath")
	}

	exec.Command("rm", localRepo).Run()

	// Sad path (to bad local dir filepath)

	exec.Command("touch", localDir).Run()

	_, stderr = cpConfig(sadDestPath, false)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad dest filepath")
	}

	exec.Command("rm", localDir).Run()
}

// TODO: test for createMissingTargetDirectory

func TestGetConfigs(t *testing.T) {
	// Happy path

	chdir(ConfigsSrc)
	gitStashBegin()

	statusErrors, pullErrors := getConfigs()

	if len(statusErrors) != 0 {
		t.Error("There were git status errors", statusErrors)
	}

	if len(pullErrors) != 0 {
		t.Error("There were git pull errors", pullErrors)
	}

	gitStashEnd()

	// Sad path

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

		cleanWorkingTree()
		gitStashEnd()
	} else {
		// TODO: how do we handle sad path when there are no upstream changes???
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

func TestGetLocalDirIndex(t *testing.T) {
	index := getLocalDirIndex()

	if OS == "darwin" {
		if index != 0 {
			t.Error("Incorrect index returned for Mac devices")
		}
	} else if OS == "linux" {
		if index != 1 {
			t.Error("Incorrect index returned for Linux devices")
		}
	} else {
		if index != 2 {
			t.Error("Incorrect index returned for Other devices")
		}
	}
}

func TestGetOSSpecificDestinationPath(t *testing.T) {
	path := getOSSpecificDestionationPath(Ghostty)

	if path != Ghostty.localDir[getLocalDirIndex()] {
		t.Errorf("Path received (%s) was not as expected (%s)", path, Ghostty.localDir[getLocalDirIndex()])
	}
}

func TestGetRsyncPaths(t *testing.T) {
	type RsyncTest struct {
		config   Config
		expect   string
		target   string
		upstream bool
	}

	tests := []RsyncTest{
		// TEST: If copying upstream && config is a directory, dest == x
		{config: Alacritty, upstream: true, target: "dest", expect: path.Dir(Alacritty.localRepo)},
		// TEST: If copying upstream && config is not a directory, dest == x
		{config: Vim, upstream: true, target: "dest", expect: Vim.localRepo},
		// TEST: If copying upstream, src == x
		{config: Alacritty, upstream: true, target: "src", expect: getOSSpecificDestionationPath(Alacritty) + "/"},
		// TEST: If copying downstream, dest == x
		// {config: Alacritty, upstream: false, target: "dest", expect: path.Dir(getOSSpecificDestionationPath(Alacritty))},
		// TEST: If copying downstram, src == x
		{config: Alacritty, upstream: false, target: "src", expect: Alacritty.localRepo},
	}

	for _, v := range tests {
		dest, src := getRsyncPaths(v.config, v.upstream)

		if v.target == "dest" {
			if dest != v.expect {
				t.Errorf("Rsync destination path (%s) not as expected (%s)", dest, v.expect)
			}
		}

		if v.target == "src" {
			if src != v.expect {
				t.Errorf("Rsync source path (%s) not as expected (%s)", src, v.expect)
			}
		}
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
		{input: "~/dev/configs/vim", output: home + "/dev/configs/vim"},
	}

	for _, v := range tests {
		replaced := replaceTildeInPath(v.input)

		if replaced != v.output {
			t.Errorf("Tilde in path was not replaced properly; test: %s; replaced: %s", v.input, replaced)
		}
	}
}
