// Script to pull my local vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub
package configspp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func chdir(dir string) {
	cherr := os.Chdir(dir)
	if cherr != nil {
		fmt.Println("cherr", cherr)
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

func getPullDirs() []string {
	var eslint_dir, ghostty_dir, nvim_dir, stylelint_dir, vimrc_dir string

	HOME_PATH := "/home/johnsoct/"
	OS := runtime.GOOS

	if OS == "linux" {
		eslint_dir = HOME_PATH + "dev/configs/eslint/"
		ghostty_dir = HOME_PATH + "dev/configs/ghostty/"
		nvim_dir = HOME_PATH + ".config/nvim/"
		stylelint_dir = HOME_PATH + "dev/configs/stylelint/"
		vimrc_dir = HOME_PATH + "dev/configs/"
	} else if OS == "darwin" {
		eslint_dir = HOME_PATH + "dev/configs/eslint/"
		ghostty_dir = HOME_PATH + "dev/configs/ghostty/"
		nvim_dir = HOME_PATH + ".config/nvim/"
		stylelint_dir = HOME_PATH + "dev/configs/stylelint/"
		vimrc_dir = HOME_PATH + "dev/configs/"
	}

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

func main() {
	dirs := getPullDirs()

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
