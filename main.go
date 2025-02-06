// Script to pull my local vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	ConfigsSrc      = []string{getHomePath() + "/dev/configs"}
	EslintDest      = ConfigsSrc[0] + EslintSrc
	EslintSrc       = ConfigsSrc[0] + "/eslint"
	GhosttyDest     = []string{getHomePath(), getHomePath() + "/Library/Application Support/com.mitchellh.ghostty"}
	GhosttySrc      = ConfigsSrc[0] + "/ghostty"
	NvimDest        = getHomePath()
	NvimSrc         = ConfigsSrc[0] + "/nvim"
	OS              = runtime.GOOS
	StylelintDest   = ConfigsSrc[0] + StylelintSrc
	StylelintSrc    = ConfigsSrc[0] + "/stylelint"
	UncommittedText = "Changes not staged for commit:"
	VimDest         = getHomePath()
	VimSrc          = ConfigsSrc[0] + "/vim"
)

func chdir(dir string) {
	local_dir := replaceTildeInPath(dir)

	cherr := os.Chdir(local_dir)
	if cherr != nil {
		fmt.Println("Error changing directory", cherr)
	}
}

func cpConfigs() {
	ghosttyConfigName := getGhosttyConfigName()

	_, copyError := cpGhosttyConfig(GhosttySrc + "/" + ghosttyConfigName)
	if copyError != nil {
		fmt.Fprintf(os.Stderr, "Error while copying ghostty's config from /configs %v\n", copyError)
	}

	_, copyError = cpVimConfig(VimSrc + "/.vimrc")
	if copyError != nil {
		fmt.Fprintf(os.Stderr, "Error while copying .vimrc from /configs %v\n", copyError)
	}
}

func cpGhosttyConfig(filepath string) ([]byte, error) {
	destination := getGhosttyDestination()

	Stdout, Stderr := exec.Command("cp", filepath, destination).Output()

	return Stdout, Stderr
}

func cpVimConfig(filepath string) ([]byte, error) {
	Stdout, Stderr := exec.Command("cp", filepath, VimDest).Output()

	return Stdout, Stderr
}

func getConfigs() ([]error, []error) {
	pullErrors := make([]error, 0)
	statusErrors := make([]error, 0)

	for _, dir := range ConfigsSrc {
		_, statusError := getGitStatus(dir)
		if statusError != nil {
			statusErrors = append(statusErrors, statusError)
		}

		_, pullError := pullFromGit(dir)
		if pullError != nil {
			fmt.Println(pullError)
			pullErrors = append(pullErrors, pullError)
		}
	}

	return statusErrors, pullErrors
}

func getGhosttyConfigName() string {
	if OS == "darwin" {
		return "mac_config"
	} else {
		return "config"
	}
}

func getGhosttyDestination() string {
	if OS == "darwin" {
		return GhosttyDest[1]
	} else {
		return GhosttyDest[0]
	}
}

func getGitStatus(dir string) ([]byte, error) {
	chdir(dir)

	Stdout, Stderr := exec.Command("git", "status").Output()
	if Stderr != nil {
		return nil, Stderr
	}

	// NOTE: not checking for untracked files because we're rebasing on pull
	if strings.Contains(string(Stdout), UncommittedText) {
		return nil, errors.New("Changes need to be committed in" + dir)
	}

	return Stdout, nil
}

func getHomePath() string {
	// NOTE: I am not worrying about the possibility of an error because
	// none of my machines, in reality or theoretical, could operate without
	// $HOME set
	home, _ := os.UserHomeDir()

	return home
}

func gitStashBegin() {
	exec.Command("git", "stash").Run()
}

func gitStashEnd() {
	exec.Command("git", "stash", "apply").Run()
	exec.Command("git", "stash", "clear").Run()
}

func pullFromGit(dir string) ([]byte, error) {
	chdir(dir)

	gitStashBegin()

	Stdout, Stderr := exec.Command("git", "pull", "--rebase").Output()

	gitStashEnd()

	return Stdout, Stderr
}

func replaceTildeInPath(path string) string {
	local_path := path
	indexOfTilde := strings.IndexRune(local_path, '~')

	if indexOfTilde != -1 {
		local_path = getHomePath() + local_path[indexOfTilde+1:]
	}

	return local_path
}

func main() {
	statusErrors, pullErrors := getConfigs()
	if statusErrors != nil {
		fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", statusErrors)
	} else if pullErrors != nil {
		fmt.Fprintf(os.Stderr, "Error pulling from git: %v\n", pullErrors)
	}

	cpConfigs()
}
