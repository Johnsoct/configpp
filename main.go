// Script to pull my local vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// TODO: 1. Get location of configs depending on OS (mac/linux)
	// TODO: 2. Check if there are any pending changes, and if so, report them
	// TODO: 3. Pull changes from GitHub
	// TODO: 4. Source bashrc?
	var eslint_dir, ghostty_dir, nvim_dir, stylelint_dir, vimrc_dir string

	OS := os.Getenv("GOOS")

	if OS == "linux" {
		eslint_dir = "$HOME/dev/configs/eslint/"
		ghostty_dir = "$HOME/dev/configs/ghostty/"
		nvim_dir = "$HOME/.config/nvim/"
		stylelint_dir = "$HOME/dev/configs/stylelint/"
		vimrc_dir = "$HOME/dev/configs/"
	} else if OS == "darwin" {
		eslint_dir = "$HOME/dev/configs/eslint/"
		ghostty_dir = "$HOME/dev/configs/ghostty/"
		nvim_dir = "$HOME/.config/nvim/"
		stylelint_dir = "$HOME/dev/configs/stylelint/"
		vimrc_dir = "$HOME/dev/configs/"
	}

	// TODO: Check if there are uncommitted changes
	// TODO: Read the output of the command some how
	cmd := exec.Command("cd", eslint_dir, "git", "log", "--oneline")

	if err := cmd.Run(); err != nil {
		fmt.Println("ERRROR!!!!", err)
	}

	fmt.Println(exec.Command("git log", eslint_dir))
	fmt.Println(exec.Command("git log", ghostty_dir))
	fmt.Println(exec.Command("git log", nvim_dir))
	fmt.Println(exec.Command("git log", stylelint_dir))
	fmt.Println(exec.Command("git log", vimrc_dir))
	// Pull changes
}
