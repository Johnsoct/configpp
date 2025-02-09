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
	dir       bool
	localDir  []string // NOTE: Mac destination at index 0
	localRepo string
}

var (
	Alacritty = Config{
		dir:       true,
		localDir:  []string{getHomePath() + "/.config/alacritty"},
		localRepo: ConfigsSrc + "/alacritty",
	}
	ConfigsSrc = getHomePath() + "/dev/configs"
	FlagCopy   = flag.Bool("c", false, "Copy local directory configurations to ~/dev/configs/")
	Eslint     = Config{
		dir:       true,
		localDir:  []string{ConfigsSrc + "/eslint"},
		localRepo: ConfigsSrc + "/eslint",
	}
	Ghostty = Config{
		dir:       false,
		localDir:  []string{getHomePath() + "/Library/Application Support/com.mitchellh.ghostty/config", getHomePath() + "/.config/ghostty/config"},
		localRepo: ConfigsSrc + "/ghostty/config",
	}
	Nvim = Config{
		dir:       true,
		localDir:  []string{getHomePath() + "/.config/nvim"},
		localRepo: ConfigsSrc + "/nvim",
	}
	OS        = runtime.GOOS
	Stylelint = Config{
		dir:       true,
		localDir:  []string{ConfigsSrc + "/stylelint"},
		localRepo: ConfigsSrc + "/stylelint",
	}
	UncommittedText = "Changes not staged for commit:"
	Vim             = Config{
		dir:       false,
		localDir:  []string{getHomePath() + "/.vimrc"},
		localRepo: ConfigsSrc + "/vim/.vimrc",
	}
	Zellij = Config{
		dir:       true,
		localDir:  []string{getHomePath() + "/.config/zellij"},
		localRepo: ConfigsSrc + "/zellij",
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
	osSpecificDestination := config.localDir[0]
	if OS == "linux" {
		osSpecificDestination = config.localDir[1]
	}

	var dest string
	var src string
	if fromDestination {
		if config.dir {
			dest = ConfigsSrc
		} else {
			dest = config.localRepo
		}
		src = osSpecificDestination
	} else {
		dest = osSpecificDestination
		src = config.localRepo
	}

	fmt.Printf("Src: %s | Dest: %s\n", src, dest)
	rsync := exec.Command("rsync", "-arv", "--progress", src, dest, "--exclude", ".git")

	var stderr error
	var stdout []byte
	if !fromDestination {
		stdout, stderr = rsync.CombinedOutput()
	} else {
		// NOTE: cp/rsync'ing directories will create the target directory if missing
		// but cp/rsync'ing a specific file to a non-existent directory fails
		if !config.dir {
			targetDirectory := path.Dir(dest)
			_, statErr := os.Stat(targetDirectory)
			if os.IsNotExist(statErr) {
				fmt.Println(statErr)
				exec.Command("mkdir", path.Dir(config.localRepo)).Run()
			}
		}

		stdout, stderr = rsync.CombinedOutput()
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
