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
	OS = runtime.GOOS
)

func chdir(dir string) {
	local_dir, err := replaceTildeInPath(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error replacing the tilde in dir; dir: %s; error: %v", dir, err)
	}

	cherr := os.Chdir(local_dir)
	if cherr != nil {
		fmt.Println("cherr", cherr)
	}
}

func cpConfigs(dirs []string) {
	var ghostySrc string
	if OS == "linux" {
		ghostySrc = "config"
	} else if OS == "darwin" {
		ghostySrc = "mac_config"
	}

	_, ghosttyCpErr := cpGhosttyConfig(dirs[0] + ghostySrc)
	if ghosttyCpErr != nil {
		fmt.Fprintf(os.Stderr, "Error while copying ghostty's config from /configs %v\n", ghosttyCpErr)
	}

	_, vimCpErr := cpVimConfig(dirs[0] + "/vim/.vimrc")
	if vimCpErr != nil {
		fmt.Fprintf(os.Stderr, "Error while copying .vimrc from /configs %v\n", vimCpErr)
	}
}

func cpGhosttyConfig(filepath string) ([]byte, error) {
	destination := getGhosttyDestination()

	Stdout, Stderr := exec.Command("cp", filepath, destination).Output()

	return Stdout, Stderr
}

func cpVimConfig(filepath string) ([]byte, error) {
	destination := getHomePath()

	Stdout, Stderr := exec.Command("cp", filepath, destination).Output()

	return Stdout, Stderr
}

func getConfigs(dirs []string) ([]error, []error) {
	pullErrors := make([]error, 0)
	statusErrors := make([]error, 0)

	for _, dir := range dirs {
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

func getGhosttyDestination() string {
	destination := getHomePath()

	if OS == "linux" {
		destination += "/.config/ghostty/"
	} else if OS == "darwin" {
		destination += "/Library/Application Support/com.mitchellh.ghostty/"
	}

	return destination
}

func getGitStatus(dir string) ([]byte, error) {
	chdir(dir)

	Stdout, Stderr := exec.Command("git", "status").Output()
	if Stderr != nil {
		return nil, Stderr
	}

	// NOTE: not checking for untracked files because we're rebasing on pull
	uncommittedText := "Changes not staged for commit:"
	uncommittedCheck := strings.Contains(string(Stdout), uncommittedText)

	if uncommittedCheck {
		return nil, errors.New("Changes need to be committed in" + dir)
	}

	return Stdout, nil
}

func getHomePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "User's home directory could not be retrieved!")
	}

	return home
}

func getPullDirs() []string {
	homepath := getHomePath()

	configs_dir := homepath + "/dev/configs/"
	nvim_dir := homepath + "/.config/nvim/"

	return []string{
		configs_dir,
		nvim_dir,
	}
}

func pullFromGit(dir string) ([]byte, error) {
	chdir(dir)

	Stdout, Stderr := exec.Command("git", "pull", "--rebase").Output()
	if Stderr != nil {
		return nil, Stderr
	}

	return Stdout, nil
}

func replaceTildeInPath(path string) (string, error) {
	local_path := path
	indexOfTilde := strings.IndexRune(local_path, '~')

	if indexOfTilde != -1 {
		home, err := os.UserHomeDir()
		if err != nil {
			return local_path, err
		}

		local_path = home + local_path[indexOfTilde+1:]
	}

	return local_path, nil
}

func main() {
	dirs := getPullDirs()

	statusErrors, pullErrors := getConfigs(dirs)
	if statusErrors != nil {
		fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", statusErrors)
	} else if pullErrors != nil {
		fmt.Fprintf(os.Stderr, "Error pulling from git: %v\n", pullErrors)
	}

	cpConfigs(dirs)
}
