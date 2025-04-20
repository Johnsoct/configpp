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
	Bashrc = Config{
		dir: false,

		localDir:  []string{getHomePath() + "/.bashrc"},
		localRepo: ConfigsSrc + "/bash/.bashrc",
	}
	ConfigsSrc   = getHomePath() + "/dev/configs"
	FlagUpstream = flag.Bool("u", false, "Copy local directory configurations to upstream ("+ConfigsSrc+")")
	Eslint       = Config{
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
		Bashrc,
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

// Copies a directory or file from one location to another via `rsync`.
// Utilizes the current OS archiecture to discern the local directory.
// Respects whether the user passed `-c` in the CLI call to discern
// which direction to copy in.
// Utilizing path.Dir to get the last directory of a path when the path
// ends with a file, such as for Vim or Ghostty.
// Excludes .git folders in local directories when copying.
// Discerns whether the destination has the targeted directory when copying
// a file instead of a directory.
func cpConfig(config Config, upstream bool) ([]byte, error) {
	dest, src := getRsyncPaths(config, upstream)
	fmt.Printf("-----------------------------------\n")
	fmt.Printf("\nCopying [%s] to [%s]\n", src, dest)

	rsync := exec.Command("rsync", "-arv", "--progress", src, dest, "--exclude", ".git")

	if !upstream {
		return rsync.CombinedOutput()
	} else {
		fmt.Printf("\n [%s] does not exist in [%s]; creating missing directory", src, ConfigsSrc)

		// NOTE: cp/rsync'ing directories will create the target directory if missing
		// but cp/rsync'ing a specific file to a non-existent directory fails
		createMissingTargetDirectory(config, dest)

		return rsync.CombinedOutput()
	}
}

func cpConfigs() {
	for _, config := range ConfigsToCopy {
		stdout, stderr := cpConfig(config, *FlagUpstream)

		if stderr != nil {
			fmt.Fprintf(os.Stderr, "Error while copying config - %s\n", stdout)
		} else {
			fmt.Printf("\nRsync output: %s\n", stdout)
		}
	}
}

func createMissingTargetDirectory(config Config, dest string) {
	if !config.dir {
		targetDirectory := path.Dir(dest)
		_, statErr := os.Stat(targetDirectory)
		if os.IsNotExist(statErr) {
			fmt.Printf("\n%s", statErr)
			exec.Command("mkdir", path.Dir(config.localRepo)).Run()
		}
	}
}

func deleteLocalShareNvim() {
	_, stderr := exec.Command("rm", "-rf", replaceTildeInPath("~/.local/share/nvim")).CombinedOutput()

	if stderr != nil {
		fmt.Fprintf(os.Stderr, "There was an issue removing nvim's local share directory [%s]", stderr)
	}

	fmt.Println("nvim deleted from ~/.local/share/nvim")
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

func getLocalDirIndex() int {
	if OS == "darwin" {
		return 0
	} else if OS == "linux" {
		return 1
	} else {
		return 2
	}
}

func getOSSpecificDestionationPath(config Config) string {
	return config.localDir[getLocalDirIndex()]
}

func getRsyncPaths(config Config, upstream bool) (string, string) {
	destPathByOS := getOSSpecificDestionationPath(config)

	var dest string
	var src string
	if upstream {
		if config.dir {
			dest = path.Dir(config.localRepo)
		} else {
			dest = config.localRepo
		}
		src = destPathByOS
	} else {
		dest = path.Dir(destPathByOS)
		src = config.localRepo
	}

	return dest, src
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

	if *FlagUpstream {
		cpConfigs()

		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Updates to configs").Run()

	} else {
		// Pull most recent changes from upstream (git)
		statusErrors, pullErrors := getConfigs()
		if len(statusErrors) != 0 {
			fmt.Fprintf(os.Stderr, "Errors checking git status: %v\n", statusErrors)
		} else if len(pullErrors) != 0 {
			fmt.Fprintf(os.Stderr, "Errors pulling from git: %v\n", pullErrors)
		}

		cpConfigs()
		deleteLocalShareNvim()
	}
}
