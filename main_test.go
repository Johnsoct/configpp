package main

import (
	"fmt"
	"log"
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

/*
 * Executes the provided command and arguments from the provided directory.
 *
 * If an error occurs, logs fatal.
 * If successful, prints the command's output.
 */
func executeCommand(dir string, command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Dir = dir

	stdout, stderr := cmd.CombinedOutput()
	if stderr != nil {
		log.Fatalf("Error during %s %s in %s:\n%v\nError:\n%s", command, arg, dir, stderr, stdout)
	}

	fmt.Printf("%s", stdout)
}

/*
 * Adds everything in the working tree to the index from the provided directory
 */
func gitAddAll(dir string) {
	executeCommand(dir, "git", "add", ".")
}

/*
 * Hard resets the index from the provided directory
 */
func gitCleanWorkingTree(dir string) {
	executeCommand(dir, "git", "reset", "--hard")
}

/*
 * Executes a commit with a generic message from the provided directory
 */
func gitCommit(dir string) {
	executeCommand(dir, "git", "commit", "-m", "'test commit'")
}

/*
 * Creates a safe git repository to test git commands in to avoid polluting this actual repo
 * 1. Creates and initializes a remote git directory
 * 2. Creates a temporary git directory
     * 2.1 Initializes temp dir as git repo
     * 2.2 Sets local git config settings for user (email/name)
     * 2.3 Sets the origin remote to our remote git directory
     * 2.4 Calls the callback
     * 2.5 Removes temp dir
 * 3 Removes remote git dir
*/
func gitCreateSandbox(callback func(dir string)) {
	remoteDir, err := os.MkdirTemp("", "remote_git_dir")
	if err != nil {
		log.Fatal(err)
	}

	localInstallPath, err := os.MkdirTemp("", "local_git_dir")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(remoteDir)
	defer os.RemoveAll(localInstallPath)

	gitInitDirectory(remoteDir, true)
	gitInitDirectory(localInstallPath, false)
	gitLocalConfigDetails(localInstallPath)
	gitSetRemote(localInstallPath, remoteDir)

	callback(localInstallPath)
}

/*
 * Creates an empty file, test.txt, from the provided directory
 */
func gitDirtyLocalRepo(dir string) {
	executeCommand(dir, "touch", "test.txt")
}

/*
 * Initializes the provided directory as a git repository.
 *
 * For remote directories, pass `true` for `makeBare`.
 */
func gitInitDirectory(dir string, makeBare bool) {
	if makeBare {
		executeCommand(dir, "git", "init", "--bare", "-b", "main")
	} else {
		executeCommand(dir, "git", "init", "-b", "main")
	}
}

/*
 * Sets the user email and name configuration settings from the provided directory.
 *
 * Without setting, push commands fail.
 */
func gitLocalConfigDetails(dir string) {
	executeCommand(dir, "git", "config", "user.email", "test@test.com")
	executeCommand(dir, "git", "config", "user.name", "Test User")
}

/*
 * Sets a remote named origin from a local directory to a remote directory.
 */
func gitSetRemote(localPath string, remotePath string) {
	executeCommand(localPath, "git", "remote", "add", "origin", remotePath)
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

// TEST: Missing tests for cp'ing directories
// TEST: Missing tests when the target of cp/rsync is missing the directory
// being targeted (in both directions)
// Lines 121-127 of main.go
func TestCPConfig(t *testing.T) {
	baseDest := "/tmp"
	baseSrc := "/tmp/test"
	fauxFile := "dummy.txt"
	localInstallPath := baseDest + "/" + fauxFile
	localDotfilesRepoPath := baseSrc + "/" + fauxFile

	executeCommand("/tmp", "mkdir", "test")

	happyPath := Config{
		dir:                   false,
		localInstallPath:      []string{localInstallPath, localInstallPath},
		localDotfilesRepoPath: localDotfilesRepoPath,
	}
	sadSrcPath := Config{
		dir:                   false,
		localInstallPath:      []string{localInstallPath, localInstallPath},
		localDotfilesRepoPath: "~/dev/configpp/does-not-exist.txt",
	}
	sadDestPath := Config{
		dir:                   false,
		localInstallPath:      []string{localInstallPath + "badddddddd", localInstallPath + "badddddddd"},
		localDotfilesRepoPath: localDotfilesRepoPath,
	}

	/*
	 * Downstream happy path (single file to a local dir)
	 *
	 * 1. Create a file in our local repository
	 * 2. Copy the file from our local repository to the destination
	 * 3. If there is an error copying, FAIL
	 * 4. If there is an error looking for the test's destination path, FAIL
	 * 5. If there is an error looking for the file in the test's destination, FAIL
	 */

	fmt.Printf("\n\nDownstream happy path\n")

	exec.Command("touch", localDotfilesRepoPath).Run()

	stdout, stderr := cpConfig(happyPath, false)
	if stderr != nil {
		t.Errorf("There was an unexpected error copying:\n%s\n", stdout)
	}

	stdout, stderr = exec.Command("ls", path.Dir(happyPath.localInstallPath[0])).CombinedOutput()
	if stderr != nil {
		t.Errorf("There was an unexpected error checking the test's destination:\n%s\n", stdout)
	}

	if !strings.Contains(string(stdout), fauxFile) {
		t.Errorf("The faux file was not found")
	}

	exec.Command("rm", localInstallPath).Run()
	exec.Command("rm", localDotfilesRepoPath).Run()

	/*
	 * Upstream happy path (single file from a local config dir to a local repo)
	 *
	 * 1. Create a file in a local config directory
	 * 2. Copy the file from our local config directory to the destination, a local repo
	 * 3. If there is an error copying, FAIL
	 * 4. If there is an error looking for the test's destination path, FAIL
	 * 5. If there is an error looking for the file in the test's destination, FAIL
	 */

	fmt.Printf("\n\nUpstream happy path\n")

	exec.Command("touch", localInstallPath).Run()

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

	exec.Command("rm", localInstallPath).Run()
	exec.Command("rm", localDotfilesRepoPath).Run()

	/*
	 * Upstream sad path (single file from a bad local config dir to a local repo)
	 *
	 * 1. Create a file in a local repo directory
	 * 2. Copy the file from our local repo directory to the destination, a local repo
	 * 3. If there is an error copying, FAIL
	 */

	fmt.Printf("\n\nUpstream sad path\n")

	exec.Command("touch", localDotfilesRepoPath).Run()

	_, stderr = cpConfig(sadSrcPath, true)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad src filepath")
	}

	exec.Command("rm", localDotfilesRepoPath).Run()

	// Sad path (to bad local dir filepath)
	/*
	 * Downstream sad path (single file from a local repo to a bad local config directory)
	 *
	 * 1. Create a file in a local repo directory
	 * 2. Copy the file from our local repo directory to the destination, a bad local repo
	 * 3. If there is an error copying, FAIL
	 */

	fmt.Printf("\n\nDownstream sad path\n")

	exec.Command("touch", localInstallPath).Run()

	_, stderr = cpConfig(sadDestPath, false)
	if stderr == nil {
		t.Errorf("Expected an error copying with a bad dest filepath")
	}

	exec.Command("rm", localInstallPath).Run()

	executeCommand("/tmp", "rm", "-rf", "test")
}

// TEST: test for createMissingTargetDirectory
func TestCreateMissingTargetDirectory(t *testing.T) {}

// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
// TODO: the test passed... I don't trust it
func TestGetConfigs(t *testing.T) {
	// Happy path - clean working tree, so there's no chance for errors
	gitCreateSandbox(func(dir string) {
		statusErrors, pullErrors := getConfigs(dir)

		if len(statusErrors) != 0 {
			t.Error("There were git status errors", statusErrors)
		}

		if len(pullErrors) != 0 {
			t.Error("There were git pull errors", pullErrors)
		}
	})

	// Sad path - TBD???
	gitCreateSandbox(func(dir string) {
		statusErrors, pullErrors := getConfigs(dir)

		// NOTE: if there aren't changes upstream, "Already up to date" is returned
		// regardless of what we stash/change
		exec.Command("git", "fetch").Run()
		stdout, stderr := exec.Command("git", "status").Output()
		if stderr != nil {
			t.Error("Unexpected error checking git status", stderr)
		}

		if strings.Contains(string(stdout), "Your branch is behind 'origin/") {
			gitStashBegin()
			// gitDirtyWorkingTree()

			statusErrors, pullErrors = getConfigs(dir)

			if len(statusErrors) == 0 {
				t.Error("There were git status errors", statusErrors)
			}

			if len(pullErrors) == 0 {
				t.Error("There were git pull errors", pullErrors)
			}

			gitCleanWorkingTree(dir)
		} else {
			// TODO: how do we handle sad path when there are no upstream changes???
		}
	})
}

// func TestGetGitStatus(t *testing.T) {
// 	happyPath := "nothing to commit, working tree clean"
// 	path := getHomePath() + "/dev/configpp"
//
// 	// Happy path
//
// 	chdir(path)
// 	gitStashBegin()
//
// 	stdout, stderr := gitStatus(path)
// 	if stderr != nil {
// 		t.Error("There was an error calling gitStatus()", stderr)
// 	}
//
// 	stdoutString := string(stdout)
// 	if !strings.Contains(stdoutString, happyPath) {
// 		t.Errorf("Expected gitStatus() to contain [%s], but it returned [%s]", happyPath, stdoutString)
// 	}
//
// 	gitStashEnd()
//
// 	// Sad path
//
// 	gitStashBegin()
// 	// gitDirtyWorkingTree()
//
// 	_, stderr = gitStatus(path)
// 	if stderr != nil {
// 		t.Error("Unexpected gitStatus() to throw an error")
// 	}
//
// 	gitCleanWorkingTree()
// 	gitStashEnd()
// }

// func TestGetLocalDirIndex(t *testing.T) {
// 	index := getLocalDirIndex()
//
// 	if OS == "darwin" {
// 		if index != 0 {
// 			t.Error("Incorrect index returned for Mac devices")
// 		}
// 	} else if OS == "linux" {
// 		if index != 1 {
// 			t.Error("Incorrect index returned for Linux devices")
// 		}
// 	} else {
// 		if index != 2 {
// 			t.Error("Incorrect index returned for Other devices")
// 		}
// 	}
// }

// func TestGetOSSpecificDestinationPath(t *testing.T) {
// 	path := getOSSpecificDestionationPath(Ghostty)
//
// 	if path != Ghostty.localInstallPath[getLocalDirIndex()] {
// 		t.Errorf("Path received (%s) was not as expected (%s)", path, Ghostty.localInstallPath[getLocalDirIndex()])
// 	}
// }

// func TestGetRsyncPaths(t *testing.T) {
// 	type RsyncTest struct {
// 		config   Config
// 		expect   string
// 		target   string
// 		upstream bool
// 	}
//
// 	tests := []RsyncTest{
// 		// TEST: If copying upstream && config is a directory, dest == config localDotfilesRepoPath including (copying the contents of src into dest)
// 		{config: Alacritty, upstream: true, target: "dest", expect: Alacritty.localDotfilesRepoPath},
// 		// TEST: If copying upstream && config is not a directory, dest == configs root directory (copying local file to config dir root)
// 		{config: Vim, upstream: true, target: "dest", expect: Vim.localDotfilesRepoPath},
// 		// TEST: If copying upstream && config is a directory, src == config.localInstallPath + "/" (rsync only copies dir contents if dir ends "/")
// 		{config: Alacritty, upstream: true, target: "src", expect: getOSSpecificDestionationPath(Alacritty) + "/"},
// 		// TEST: If copying downstream && config is a directory, dest == config localInstallPath - "config name"
// 		{config: Alacritty, upstream: false, target: "dest", expect: path.Dir(getOSSpecificDestionationPath(Alacritty))},
// 		// TEST: If copying downstram && config is a directory, src == config's localDotfilesRepoPath in configs dir
// 		{config: Alacritty, upstream: false, target: "src", expect: Alacritty.localDotfilesRepoPath},
// 	}
//
// 	for _, v := range tests {
// 		dest, src := getRsyncPaths(v.config, v.upstream)
//
// 		if v.target == "dest" {
// 			if dest != v.expect {
// 				t.Errorf("Rsync destination path (%s) not as expected (%s)", dest, v.expect)
// 			}
// 		}
//
// 		if v.target == "src" {
// 			if src != v.expect {
// 				t.Errorf("Rsync source path (%s) not as expected (%s)", src, v.expect)
// 			}
// 		}
// 	}
// }

// func TestPullFromGit(t *testing.T) {
// 	path := getHomePath() + "/dev/configpp"
//
// 	// Happy path - clear all changes from git status
//
// 	gitStashBegin()
//
// 	_, stderr := gitPull(path)
// 	if stderr != nil {
// 		t.Errorf("There was an unexpected error from gitPull(): [%s]", stderr.Error())
// 	}
//
// 	gitStashEnd()
//
// 	// Sad path - create uncommitted changes to dirty the working tree to prevent pulling
//
// 	gitStashBegin()
// 	// gitDirtyWorkingTree()
//
// 	_, stderr = gitPull(path)
// 	if stderr != nil {
// 		t.Errorf("Error expected from gitPull()")
// 	}
//
// 	gitCleanWorkingTree()
// 	gitStashEnd()
// }

func TestPushToGit(t *testing.T) {
	gitCreateSandbox(func(dir string) {
		gitDirtyLocalRepo(dir)
		gitAddAll(dir)
		gitCommit(dir)

		_, stderr := gitPush(dir)
		if stderr != nil {
			t.Errorf("Git Push failed; %v", stderr)
		}
	})
}

// func TestGetHomePath(t *testing.T) {
// 	if OS == "darwin" {
// 		expect := "/Users/" + USERNAME
// 		if getHomePath() != expect {
// 			t.Errorf("Homepath returned for Mac is incorrect")
// 		}
// 	} else {
// 		expect := "/home/" + USERNAME
// 		if getHomePath() != expect {
// 			t.Errorf("Homepath returned for Linux is incorrect")
// 		}
// 	}
// }

// func TestReplaceTildeInPath(t *testing.T) {
// 	home := getHomePath()
// 	tests := []InputOutput{
// 		{input: "~/dev", output: home + "/dev"},
// 		{input: "~/dev/configs/vim", output: home + "/dev/configs/vim"},
// 	}
//
// 	for _, v := range tests {
// 		replaced := replaceTildeInPath(v.input)
//
// 		if replaced != v.output {
// 			t.Errorf("Tilde in path was not replaced properly; test: %s; replaced: %s", v.input, replaced)
// 		}
// 	}
// }
