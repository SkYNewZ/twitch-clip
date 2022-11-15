
<a name="v0.0.2"></a>
## v0.0.2

> 2021-12-13

### Feat

* Use Taskfile
* Windows artifact is okay
* Refacto assets and make a Windows build
* Make player and streamlink public packages
* Make a real player package
* Stop using global vars
* **app:** Open Twitch button
* **notification:** Implement notifications on macOS
* **windows:** Setup notification

### Fix

* Various fixes
* Some refactoring
* Refacto streamlink client
* Refacto main app
* Logger issues on Windows
* Upgrade to Go 1.17
* Create golangci-lint config file
* Do not stop routines when error occured
* Make logger on both syslog and stdout
* PATH issues
* Some refactoring
* PATH workaround only for unix distros
* **display:** Use user login if username not received yet
* **logger:** Do not include syslog on incompatible distros
* **notification:** Run server only if supported
* **notification:** Trigger notification if set in config file
* **windows:** Images are now displayed

