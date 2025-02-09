// Script to pull my alacritty, zellij, vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub to ~/dev/configs (src) and copy them to their local destinations
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

type Config struct {
	// NOTE: Mac destination at index 0
	dest []string
	src  string
}

var (
	Alacritty = Config{
		dest: []string{getHomePath() + "/.config/alacritty"},
		src:  ConfigsSrc + "/alacritty",
	}
	ConfigsSrc = getHomePath() + "/dev/configs"
	FlagCopy   = flag.Bool("c", false, "Copy local directory configurations to ~/dev/configs/")
	Eslint     = Config{
		dest: []string{ConfigsSrc + "/eslint"},
		src:  ConfigsSrc + "/eslint",
	}
	Ghostty = Config{
		dest: []string{getHomePath() + "/Library/Application Support/com.mitchellh.ghostty", getHomePath() + "/.config/ghostty"},
		src:  ConfigsSrc + "/ghostty",
	}
	Nvim = Config{
		dest: []string{getHomePath() + "/.config/nvim"},
		src:  ConfigsSrc + "/nvim",
	}
	OS        = runtime.GOOS
	Stylelint = Config{
		dest: []string{ConfigsSrc + "/stylelint"},
		src:  ConfigsSrc + "/stylelint",
	}
	UncommittedText = "Changes not staged for commit:"
	Vim             = Config{
		dest: []string{getHomePath() + "/.vimrc"},
		src:  ConfigsSrc + "vim/.vimrc",
	}
	Zellij = Config{
		dest: []string{getHomePath() + "/.config/zellij"},
		src:  ConfigsSrc + "/zellij",
	}

	Configs = []Config{
		Alacritty,
		Eslint,
		Ghostty,
		Nvim,
		Stylelint,
		Vim,
		Zellij,
	}
	ConfigsToCopy = []Config{
		Alacritty,
		Ghostty,
		Nvim,
		Vim,
		Zellij,
	}
)

func chdir(dir string) {
	local_dir := replaceTildeInPath(dir)

	cherr := os.Chdir(local_dir)
	if cherr != nil {
		fmt.Fprintf(os.Stderr, "Error changing directory: [%v]", cherr)
	}
}

func cpConfig(config Config, fromDestination bool) ([]byte, error) {
	destination := config.dest[0]
	if OS == "linux" {
		destination = config.dest[1]
	}

	// TODO: ghostty and vim still not copying correctly
	rsyncToDestination := exec.Command("rsync", "-arv", "--progress", config.src, path.Dir(destination), "--exclude", ".git")
	rsyncToSrc := exec.Command("rsync", "-arv", "--progress", destination, path.Dir(config.src), "--exclude", ".git")

	var stderr error
	var stdout []byte
	if !fromDestination {
		stdout, stderr = rsyncToDestination.CombinedOutput()
	} else {
		target := path.Dir(config.src)

		_, statErr := os.Stat(target)
		if os.IsNotExist(statErr) {
			fmt.Println(statErr)
			// exec.Command("mkdir", path.Dir(config.src)).Run()
		}

		// stdout, stderr = exec.Command("cp", "-rv", destination, replaceTildeInPath(config.src)).CombinedOutput()
		stdout, stderr = rsyncToSrc.CombinedOutput()
	}

	return stdout, stderr
}

func cpConfigs() {
	for _, config := range ConfigsToCopy {
		stdout, stderr := cpConfig(config, *FlagCopy)

		if stderr != nil {
			fmt.Fprintf(os.Stderr, "Error while copying config - %s\n", stdout)
		}
	}
}

func getConfigs() ([]error, []error) {
	pullErrors := make([]error, 0)
	statusErrors := make([]error, 0)

	_, statusError := getGitStatus(ConfigsSrc)
	if statusError != nil {
		statusErrors = append(statusErrors, statusError)
	}

	_, pullError := pullFromGit(ConfigsSrc)
	if pullError != nil {
		pullErrors = append(pullErrors, pullError)
	}

	return statusErrors, pullErrors
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
	flag.Parse()

	if *FlagCopy {
		// NOTE: Copy from local destinations to ~/dev/configpp/
		cpConfigs()
	} else {
		statusErrors, pullErrors := getConfigs()
		if len(statusErrors) != 0 {
			fmt.Fprintf(os.Stderr, "Errors checking git status: %v\n", statusErrors)
		} else if len(pullErrors) != 0 {
			fmt.Fprintf(os.Stderr, "Errors pulling from git: %v\n", pullErrors)
		}

		// NOTE: Copy from ~/dev/configpp/ to local destinations
		cpConfigs()
	}
}
