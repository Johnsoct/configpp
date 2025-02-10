# Config Push/Pull

After pulling and copying config files from various git repos and directories between my mac and linux machine over a hundred times, I decided to just write a go program that pulls down my configs from their respective local directories (respective of operating systems) and then copy those config files to their proper destinations.

**Update (20250207)** 
I don't want this app to push changes because I would prefer not to push generic messages for local changes or to build a CLI that asked you what you wanted to do and allowed you to commit each directory to git from this application.

However, I don't think there's an issue with copying our local config to /dev/config so I can manually call the git CLI from within that directory to commit changes to git.

## Example

I use Ghostty as my terminal, and vim/Nvim for the majority of my code editing; however, Ghostty stores its config in different places on Mac and Linux, and I didn't want to create a git repo in `~/Library/Application Support/com.mitchellh.ghostty/`, so I am storing two versions of my Ghostty config in `~/dev/configs/ghostty/`, which is kept updated in GitHub, and then after pulling those configs down, I copy them to their respective locations.

## Missing features

- [ ] Everything from [Review after "v1"](#review-after-"v1")
- [x] Happy/Sad path logic needs implemented in the actual program functions instead of implementing the state for happy path and testing an error is not returned and implementing the state for sad path and testing an error is returned... We should always test that an error isn't returned because the actual function we're testing is handling the sad path cases
- [x] Nvim needs to at some point be copied from ~/.config to ~/dev/configs/nvim
- [x] Ghostty needs to at some point be copied from ~/GhosttyDest to ~/dev/configs/ghostty
- [x] Vimrc needs to at some point be copied from ~/ to ~/dev/configs

## Review after "v1"

- [x] Tests have a lot of holes... There are a lot of possible errors from running third-party CLI commands, including Git, that I chose to ignore just to get some coverage of v1, but I'm very aware the tests could be a lot more thorough, especially at setting up initial state and testing sad path resolution
- [ ] I'm not actually testing the exact files copied were copied because I don't want to run tests that actually manipulate those files, so I'm using the same functions but manipulating the filepaths and running exec.Command to create test files to copy and then assert on
    - [ ] I could use the actual files, but I'd need to store the entire files in memory, remove them from their destinations, perform the copy, assert the existence and file contents match the file in memory
    - [ ] There'd have to be some failsafe, I think... However, the files are technically being copied and pulled from Git, where they'd be safe
- [x] The one refactor I'd genuinely like to do is to change how paths were defined and configured so there wasn't confusion on whether one function returned paths with "/" at the end or not, such as `getPullDirs()` returning without "/" and then making the mistakes of trying to build path strings but forgetting to add "/" before the appending portion

## Notes to self

- When copying (or "rsyncing") directories, the destination path should not end with the directory you're copying.
- You won't build something the most clearly, efficiently as possible on the first try. I probably changed the shape of the global variables in `main.go` three or four times before finding the least confusing, supportive structure.
- When using `~` in `exec.Command`, Go does not unravel the meaning of the tilde; however, `os.UserHomeDir()` is the standard package version of the tilde for paths.
- You cannot exclude files or subdirectories from the operation with `cp`, so I switched to `rsync`, which is a more verbose, powerful improvement.
- The best places I found to make abstractions were:
    - Where ease of writing test was increased
    - A "parent" function became more digestible and readable
- `exec.CombinedOutput()` "combines" the text-based error messages from failed/errornous commands into stdout so you can access []byte values representing CLI outputs and errors.
- `stat` is pretty handy, but specifically how it can tell you whether a ... "stat'ed" is a file or directory.
- Untracked files can be ignored with rebasing
- "Upstream" and "downstream" were much better terms than "source" and "destination" in terms of copying because they retain understanding even if swapping the direction being copied in (ex. destination to source).
- Reusing as many global functions and variables from the file you're testing saves a lot of time and bloat that could lead to innaccurate, problematic tests (i.e. writing custom Config variables vs using the ones defined in `main.go`).
- Naming is important, which I knew, but man, once I changed key names from dest/src to localDir/localRepo, I was able to keep track of my own logic with a lot less effort.
- Sometimes I think I'm a genius, but before long, my own code reminds me I am not.
