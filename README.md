# pokesay-go

Print pokemon in the CLI! An adaptation of the classic "cowsay"

- [pokesay-go](#pokesay-go)
  - [One-line installs](#one-line-installs)
  - [How it works](#how-it-works)
  - [TODO](#todo)
  - [Building binaries](#building-binaries)
    - [In docker](#in-docker)
    - [On your host OS](#on-your-host-os)

---

![pokesay-demo-3](https://user-images.githubusercontent.com/9894426/160840120-f45be867-7427-4f19-862c-46e6a4b396c3.png)

---

## One-line installs

<details>
  <summary>Click to expand!</summary>

- OSX / darwin
    ```shell
    bash -c "$(curl https://raw.githubusercontent.com/tmck-code/pokesay-go/master/scripts/install.sh)" bash darwin amd64
    ```
- OSX / darwin (M1)
    ```shell
    bash -c "$(curl https://raw.githubusercontent.com/tmck-code/pokesay-go/master/scripts/install.sh)" bash darwin arm64
    ```
- Linux x64
    ```shell
    bash -c "$(curl https://raw.githubusercontent.com/tmck-code/pokesay-go/master/scripts/install.sh)" bash linux amd64
    ```
- Android ARM64 (Termux)
    ```shell
    bash -c "$(curl https://raw.githubusercontent.com/tmck-code/pokesay-go/master/scripts/install.sh)" bash android arm64
    ```
</details>

---

## How it works

This project extends on the original `fortune | cowsay`, a simple command combo that can be added to
your .bashrc to give you a random message spoken by a cow every time you open a new shell.

```
 ☯ ~ fortune | cowsay
 ______________________________________
/ Hollywood is where if you don't have \
| happiness you send out for it.       |
|                                      |
\ -- Rex Reed                          /
 --------------------------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

These pokemon sprites used here are sourced from the awesome repo
[msikma/pokesprite](https://github.com/msikma/pokesprite)

![sprits](https://github.com/msikma/pokesprite/raw/master/resources/images/banner_gen8_2x.png)

All of these sprites are converted into a form that can be rendered in a terminal (unicode
characters and colour control sequences) by the `img2xterm` tool, found at
[rossy/img2xterm](https://github.com/rossy/img2xterm)

The last pre-compile step is to use `encoding/gob` and `go:embed` to generate a go source code file
that encodes all of the converted unicode sprites as gzipped text.

Finally, this is built with the main CLI logic in `pokesay.go` into an single executable that can be
easily popped into a directory in the user's $PATH

If all you are after is installing the program to use, then there are no dependencies required!
Navigate to the Releases and download the latest binary.

## TODO

- Short-term
  - [x] Fix bad whitespace stripping when building assets
  - [ ] List all names
- Longer-term
  - [x] Make data structure to hold categories, names and pokemon
  - [x] Increase speed
  - [ ] Improve categories to be more specific than shiny/regular

## Building binaries

### In docker

_Dependencies:_ `docker`

In order to re/build the binaries from scratch, along with all the cowfile conversion, use the handy
Makefile tasks

```shell
cd build && make build/docker build/assets build/release
```

This will produce 4 executable bin files inside the `build/bin` directory, and a heap of binary asset files in `build/assets`.

### On your host OS

You'll have to install a golang version that matches the go.mod, and ensure that other package
dependencies are installed (see the dockerfile for the dependencies)

```shell
# Generate binary asset files from the cowfiles
./build/build_assets.sh

# Finally, build the pokesay tool (this builds and uses the build/pokedex.gob file automatically)
go build pokesay.go
```

## Dependencies

The real dependencies here are the amazing repos/databases of this pokesprite that's hosted by a few different places

- gen 7 & 8: https://www.pokencyclopedia.info/en/index.php?id=sprites
- gen 1-5:   https://veekun.com/dex/downloads
