# Gettit install guide

## Building from source

Requirements :

Go compiler https://go.dev/doc/install

ffmpeg https://ffmpeg.org/download.html

GNU make https://www.gnu.org/software/make/ (Dont bother too much it should come preinstalled on most linux distros)

You should build from source if :

- You want to make sure your binary was actually built from the source code on the repo

- The binary doesn't work on your OS/Distro.

- You want a 100% up to date build.


Step 1, clone the repository which contains the source code(aka download it).

`git clone https://github.com/Solirs/Gedditsave`

Step 2, cd into the downloaded directory.

`cd gettit`

Step 3, run the install command.

`sudo make install`

Step 4, test the program.

`gettit -h`


## Distro specific package

Coming soon.

## Release tarball

Requirements :

ffmpeg https://ffmpeg.org/download.html

GNU make https://www.gnu.org/software/make/ (Again don't bother it comes with most distros)

You should use a release tarball if:

- You want a quick and easy way to install gettit

- There is no package for your distro


Step 1, download the [latest](https://github.com/Solirs/gettit/releases/tag/v1.1.0) release tarball.

Go to the releases on the github repositry and download the release.

Step 2, extract the tarball and cd into the directory containing the files

`tar -xf Gettit-release-*.tar.gz`

`cd Gettit-release-*`

Step 3, run the install command.

`sudo make install`

Step 4, test.

`gettit -h`
