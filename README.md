# twitch-clip
Twitch-aware application to watch followed live streams in your favorite media player, with a simple click

![Imgur](https://i.imgur.com/FDXwa3T.png)

## Build (from macOS)

You can use https://taskfile.dev to use predefined tasks.

### For macOS

```shell
$ task build:darwin # will build the Go binary
$ task package:darwin # package/bundle as macOS app
```

### For Windows

````shell
$ brew install mingw-w64 # Compile https://github.com/getlantern/systray
$ task build:windows
````

### For Unix

Not supported yet. I didn't find the fix for
```shell
$ task build:linux
task: [mod] go mod download
task: Task "generate" is up to date
task: [build:linux] go build -ldflags="-s -w -X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}'" -tags production -o "out/twitch_clip_${GOOS}_${GOARCH}.exe" .
# runtime/cgo
linux_syscall.c:67:13: error: implicit declaration of function 'setresgid' is invalid in C99 [-Werror,-Wimplicit-function-declaration]
linux_syscall.c:67:13: note: did you mean 'setregid'?
/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/unistd.h:593:6: note: 'setregid' declared here
linux_syscall.c:73:13: error: implicit declaration of function 'setresuid' is invalid in C99 [-Werror,-Wimplicit-function-declaration]
linux_syscall.c:73:13: note: did you mean 'setreuid'?
/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/unistd.h:595:6: note: 'setreuid' declared here
task: Failed to run task "build:linux": exit status 2

```