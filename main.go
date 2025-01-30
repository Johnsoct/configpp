// Script to pull my local vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func getGitStatus(dir string) []byte {
	cherr := os.Chdir(dir)
	if cherr != nil {
		fmt.Println("cherr", cherr)
	}

	Stdout, Stderr := exec.Command("git", "status").Output()

	if Stderr != nil {
		fmt.Println("ERRROR!!!!", Stderr)
	}

	return Stdout
}

func main() {
	// TODO: 1. Get location of configs depending on OS (mac/linux)
	// TODO: 2. Check if there are any pending changes, and if so, report them
	// TODO: 3. Pull changes from GitHub
	// TODO: 4. Source bashrc?
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

	dirs := []string{
		eslint_dir,
		ghostty_dir,
		nvim_dir,
		stylelint_dir,
		vimrc_dir,
	}

	for _, val := range dirs {
		// TODO: Check if there are uncommitted changes
		// TODO: Read the output of the command some how
		out := getGitStatus(val)

		fmt.Printf("%s\n", out)
	}
}
