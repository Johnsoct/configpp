# Config Push/Pull

After pulling and copying config files from various git repos and directories between my mac and linux machine over a hundred times, I decided to just write a go program that pulls down my configs from their respective local directories (respective of operating systems) and then copy those config files to their proper destinations.

## Example

I use Ghostty as my terminal, and vim/Nvim for the majority of my code editing; however, Ghostty stores its config in different places on Mac and Linux, and I didn't want to create a git repo in `~/Library/Application Support/com.mitchellh.ghostty/`, so I am storing two versions of my Ghostty config in `~/dev/configs/ghostty/`, which is kept updated in GitHub, and then after pulling those configs down, I copy them to their respective locations.

## Missing features

- [ ] Pass a flag to choose between pulling from GitHub and pushing to GitHub (plus copying configs to their final destination for each differing git operation)

## Review after "v1"

- [ ] Tests have a lot of holes... There are a lot of possible errors from running third-party CLI commands, including Git, that I chose to ignore just to get some coverage of v1, but I'm very aware the tests could be a lot more thorough, especially at setting up initial state and testing sad path resolution
- [ ] I'm not actually testing the exact files copied were copied because I don't want to run tests that actually manipulate those files, so I'm using the same functions but manipulating the filepaths and running exec.Command to create test files to copy and then assert on
    - [ ] I could use the actual files, but I'd need to store the entire files in memory, remove them from their destinations, perform the copy, assert the existence and file contents match the file in memory
    - [ ] There'd have to be some failsafe, I think... However, the files are technically being copied and pulled from Git, where they'd be safe
- [ ] The one refactor I'd genuinely like to do is to change how paths were defined and configured so there wasn't confusion on whether one function returned paths with "/" at the end or not, such as `getPullDirs()` returning without "/" and then making the mistakes of trying to build path strings but forgetting to add "/" before the appending portion
