# Config Push/Pull

After pulling and copying config files from various git repos and directories between my mac and linux machine over a hundred times, I decided to just write a go program that pulls down my configs from their respective local directories (respective of operating systems) and then copy those config files to their proper destinations.

## Example

I use Ghostty as my terminal, and vim/Nvim for the majority of my code editing; however, Ghostty stores its config in different places on Mac and Linux, and I didn't want to create a git repo in `~/Library/Application Support/com.mitchellh.ghostty/`, so I am storing two versions of my Ghostty config in `~/dev/configs/ghostty/`, which is kept updated in GitHub, and then after pulling those configs down, I copy them to their respective locations.

## Missing features

- [ ] Pass a flag to choose between pulling from GitHub and pushing to GitHub (plus copying configs to their final destination for each differing git operation)
