# twitch-clip
(Work in progress) Twitch-aware application to watch followed live streams in your favorite media player, with a simple click

![Imgur](https://i.imgur.com/FDXwa3T.png)

## Build (from macOS)

### For macOS

```shell
$ go build .
```

### For Windows

````shell
$ brew install mingw-w64 # Compile https://github.com/getlantern/systray
$ CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" .
````

### For Unix

Not supported yet. I didn't find the fix for
```shell
$ CGO_ENABLED=1 GOOS=linux go build .
# github.com/getlantern/systray
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:78:2: undefined: nativeLoop
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:106:2: undefined: registerSystray
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:111:14: undefined: quit
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:136:2: undefined: addSeparator
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:190:2: undefined: hideMenuItem
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:195:2: undefined: showMenuItem
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:220:2: undefined: addOrUpdateMenuItem

$ GOOS=linux go build .                                                                                      
# github.com/getlantern/systray
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:78:2: undefined: nativeLoop
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:106:2: undefined: registerSystray
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:111:14: undefined: quit
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:136:2: undefined: addSeparator
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:190:2: undefined: hideMenuItem
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:195:2: undefined: showMenuItem
../../../../pkg/mod/github.com/getlantern/systray@v1.1.0/systray.go:220:2: undefined: addOrUpdateMenuItem
```