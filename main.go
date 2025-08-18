// Script to pull my alacritty, zellij, vimrc, nvim, ghostty, prettier, stylelint, and eslint
// configs from GitHub to ~/dev/configs (src) and copy them to their local destinations
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

/*
 * Config
 *
 * `localInstallPath` represents the local config directories, such as "~/.config/alacritty." Unlike `localDotfilesRepoPath`, it is a slice because there may be different paths for the same config depending on whether the OS is Mac OSX or Linux.
 * `localDotfilesRepoPath` represents the local directory where all my dotfile directories are stored, which is typically ~/dev/configs/ + config.
 */
type Config struct {
	dir                   bool
	localInstallPath      []string
	localDotfilesRepoPath string
}

var (
	Alacritty = Config{
		dir:                   true,
		localInstallPath:      []string{getHomePath() + "/.config/alacritty"},
		localDotfilesRepoPath: ConfigsSrc + "/alacritty",
	}
	Bashaliases = Config{
		dir:                   false,
		localInstallPath:      []string{getHomePath() + "/.bash_aliases"},
		localDotfilesRepoPath: ConfigsSrc + "/bash/.bash_aliases",
	}
	Bashrc = Config{
		dir: false,

		localInstallPath:      []string{getHomePath() + "/.bashrc"},
		localDotfilesRepoPath: ConfigsSrc + "/bash/.bashrc",
	}
	ConfigsSrc = getHomePath() + "/dev/configs"
	Eslint     = Config{
		dir:                   true,
		localInstallPath:      []string{ConfigsSrc + "/eslint"},
		localDotfilesRepoPath: ConfigsSrc + "/eslint",
	}
	FlagUpstream = flag.Bool("u", false, "Copy local directory configurations to upstream ("+ConfigsSrc+")")
	FontPatcher  = Config{
		dir:                   true,
		localInstallPath:      []string{getHomePath() + "/dev/FontPatcher"},
		localDotfilesRepoPath: ConfigsSrc + "/fontpatcher",
	}
	Ghostty = Config{
		dir:                   true,
		localInstallPath:      []string{getHomePath() + "/Library/Application Support/com.mitchellh.ghostty", getHomePath() + "/.config/ghostty"},
		localDotfilesRepoPath: ConfigsSrc + "/ghostty",
	}
	Nvim = Config{
		dir:                   true,
		localInstallPath:      []string{getHomePath() + "/.config/nvim"},
		localDotfilesRepoPath: ConfigsSrc + "/nvim",
	}
	OS        = runtime.GOOS
	Stylelint = Config{
		dir:                   true,
		localInstallPath:      []string{ConfigsSrc + "/stylelint"},
		localDotfilesRepoPath: ConfigsSrc + "/stylelint",
	}
	UncommittedText = "Changes not staged for commit:"
	Vim             = Config{
		dir:                   false,
		localInstallPath:      []string{getHomePath() + "/.vimrc"},
		localDotfilesRepoPath: ConfigsSrc + "/vim/.vimrc",
	}
	Zellij = Config{
		dir:                   true,
		localInstallPath:      []string{getHomePath() + "/.config/zellij"},
		localDotfilesRepoPath: ConfigsSrc + "/zellij",
	}

	Configs = []Config{
		Alacritty,
		Bashaliases,
		Bashrc,
		Eslint,
		FontPatcher,
		Ghostty,
		Nvim,
		Stylelint,
		Vim,
		Zellij,
	}
)

/*
 * Sets the CWD to the provided directory.
 */
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
		// NOTE: cp/rsync'ing directories will create the target directory if missing
		// but cp/rsync'ing a specific file to a non-existent directory fails
		createMissingTargetDirectory(config, dest)

		return rsync.CombinedOutput()
	}
}

func cpConfigs() {
	for _, config := range Configs {
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
			fmt.Printf("\n [%s] does not exist in [%s]; creating missing directory", dest, ConfigsSrc)
			fmt.Printf("\n%s", statErr)
			if stderr := exec.Command("mkdir", path.Dir(config.localDotfilesRepoPath)).Run(); stderr != nil {
				fmt.Printf("\nThere was an error executing `mkdir %s`\n", path.Dir(config.localDotfilesRepoPath))
			}
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

func getHomePath() string {
	// NOTE: I am not worrying about the possibility of an error because
	// none of my machines, in reality or theoretical, could operate without
	// $HOME set
	home, _ := os.UserHomeDir()

	return home
}

func getLocalDirIndex() int {
	switch OS {
	case "darwin":
		return 0
	case "linux":
		return 1
	default:
		return 2
	}
}

func getOSSpecificDestionationPath(config Config) string {
	if len(config.localInstallPath) == 1 {
		return config.localInstallPath[0]
	} else {
		return config.localInstallPath[getLocalDirIndex()]
	}
}

/*
 * Returns the dest and src paths for a `rsync` operation.
 *
 * A "downstream" operation is when we pull from github, which correlates to:
 * `config.localDotfilesRepoPath` is our source path. This is our local Git repo of dotfile directories (~/dev/configs/).
 * `config.localInstallPath` is our destination path. This is our local config directories (~/.config/alacritty/)
 *
 * An "upstream" operation is when we push from our local Git repo of dotfile directories to GitHub:
 * `config.localDotfilesRepoPath` is our destination path. This is our local Git repo of dotfile directories (~/dev/configs/).
 * `config.localInstallPath` is our source path; however, if `localInstallPath` is a directory, verse a single file, we append "/" since `rsync` only copies files inside of a directory if the path ends in "/". This is our local config directories (~/.config/alacritty/).
 */
func getRsyncPaths(config Config, upstream bool) (string, string) {
	destPathByOS := getOSSpecificDestionationPath(config)

	var dest string
	var src string
	if upstream {
		dest = config.localDotfilesRepoPath
		if config.dir {
			// NOTE: rsync will only copy the files inside of a directory if the source path
			// ends in "/"
			src = destPathByOS + "/"
		} else {
			src = destPathByOS
		}
	} else {
		// dest should be the directory containing our target since we'll be overwriting it
		dest = path.Dir(destPathByOS)
		src = config.localDotfilesRepoPath
	}

	return dest, src
}

/*
 * Within the provided directory:
 * 1. Stash the current working tree changes
 * 2. Pull via rebase
 * 3. Apply the latest stash
 */
func gitPull(dir string) ([]byte, error) {
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = dir

	gitStashBegin()

	stdout, stderr := cmd.CombinedOutput()

	gitStashEnd()

	return stdout, stderr
}

/*
 * Pushes to "origin" remote's "main" branch and prints the output.
 */
func gitPush(dir string) ([]byte, error) {
	fmt.Printf("\nUpdating Git from %s\n", dir)

	cmd := exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = dir

	stdout, stderr := cmd.CombinedOutput()
	if stderr != nil {
		fmt.Printf("\n%v\n", stderr)
	}

	fmt.Printf("%s", stdout)

	return stdout, stderr
}

/*
 * Prints the output of calling `git status`
 * Returns
 */
func gitStatus(dir string) ([]byte, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = dir

	stdout, stderr := cmd.CombinedOutput()
	if stderr != nil {
		fmt.Printf("\n%v\n", stderr)
	}

	fmt.Printf("%s", stdout)

	return stdout, stderr
}

/*
 * Within the CWD, stashes the current working tree.
 */
func gitStashBegin() {
	if stderr := exec.Command("git", "stash").Run(); stderr != nil {
		fmt.Printf("\nThere was an error executing `git stash`\n")
	}
}

/*
 * Within the CWD, applies the latest stash.
 */
func gitStashEnd() {
	if stderr := exec.Command("git", "stash", "apply").Run(); stderr != nil {
		fmt.Printf("\nThere was an error executing `git stash apply`\n")
	}

	if stderr := exec.Command("git", "stash", "clear").Run(); stderr != nil {
		fmt.Printf("\nThere was an error executing `git stash clear`\n")
	}
}

/*
 * Pulls in changes from origin/<currentBranch>
 * Returns error of git status and pull
 */
func pullDownConfigs(dir string) (error, []byte) {
	pullStdout, pullStderr := gitPull(dir)
	if pullStderr != nil {
		fmt.Printf("\n%s\n%v\n", pullStdout, pullStderr)
	}

	return pullStderr, pullStdout
}

/*
 * Replaces a "~" in a path with your local $HOME path variable value.
 */
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

		if _, stderr := gitPush(ConfigsSrc); stderr != nil {
			fmt.Fprintf(os.Stderr, "Error pushing to git: %v\n", stderr)
		}
	} else {
		// Pull most recent changes from upstream (git)
		pullStderr, _ := pullDownConfigs(ConfigsSrc)
		if pullStderr != nil {
			fmt.Fprintf(os.Stderr, "Errors pulling from git: %v\n", pullStderr)
		}

		cpConfigs()
		deleteLocalShareNvim()
	}
}
