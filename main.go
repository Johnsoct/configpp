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
	var configFilepath string
	if OS == "linux" {
		configFilepath = "config"
	} else if OS == "darwin" {
		configFilepath = "mac_config"
	}

	_, ghosttyCpErr := cpGhostyConfig(dirs[1] + configFilepath)
	if ghosttyCpErr != nil {
		fmt.Fprintf(os.Stderr, "Error while copying ghostty's config from /configs %v\n", ghosttyCpErr)
	}

	_, vimCpErr := cpVimConfig(dirs[4] + ".vimrc")
	if vimCpErr != nil {
		fmt.Fprintf(os.Stderr, "Error while copying .vimrc from /configs %v\n", vimCpErr)
	}
}

func cpGhostyConfig(filepath string) ([]byte, error) {
	var destination string
	homepath := getHomePath()

	if OS == "linux" {
		destination = homepath + ".config/ghostty/config"
	} else if OS == "darwin" {
		destination = homepath + "Library/Application Support/com.mitchellh.ghostty/config"
	}

	Stdout, Stderr := exec.Command("cp", filepath, destination).Output()
	if Stderr != nil {
		return nil, Stderr
	}

	return Stdout, nil
}

func cpVimConfig(filepath string) ([]byte, error) {
	destination := getHomePath()

	Stdout, Stderr := exec.Command("cp", filepath, destination).Output()
	if Stderr != nil {
		return nil, Stderr
	}

	return Stdout, nil
}

func getConfigs(dirs []string) {
	// fmt.Println(dirs)
	for _, dir := range dirs {
		_, statusError := getGitStatus(dir)
		if statusError != nil {
			fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", statusError)
		}

		_, pullError := pullFromGit(dir)
		if pullError != nil {
			fmt.Fprintf(os.Stderr, "Error pulling from git: %v\n", pullError)
		}
	}
}

func getGitStatus(dir string) ([]byte, error) {
	chdir(dir)

	Stdout, Stderr := exec.Command("git", "status").Output()
	if Stderr != nil {
		return nil, Stderr
	}

	uncommittedText := "Changes not staged for commit:"
	if strings.Contains(string(Stdout), uncommittedText) {
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
	var eslint_dir, ghostty_dir, nvim_dir, stylelint_dir, vimrc_dir string
	homepath := getHomePath()

	eslint_dir = homepath + "/dev/configs/eslint/"
	ghostty_dir = homepath + "/dev/configs/ghostty/"
	nvim_dir = homepath + "/.config/nvim/"
	stylelint_dir = homepath + "/dev/configs/stylelint/"
	vimrc_dir = homepath + "/dev/configs/vim/"

	return []string{
		eslint_dir,
		ghostty_dir,
		nvim_dir,
		stylelint_dir,
		vimrc_dir,
	}
}

func pullFromGit(dir string) ([]byte, error) {
	chdir(dir)

	Stdout, Stderr := exec.Command("git", "pull").Output()
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

	getConfigs(dirs)
	cpConfigs(dirs)
}
