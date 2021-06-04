package player

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var _ Player = (*player)(nil)

// Player describes an available media player application
// If you want to use a custom one, make sure to implement this interface
type Player interface {
	// Name return the current player name
	Name() string

	// Run process current URL through current player
	// u will be the stream URL
	// title will be the stream title
	Run(u, title string, output io.Writer) error
}

type player struct {
	name       string
	command    []string
	registry   string
	registry32 string
}

func (p *player) Name() string {
	return p.name
}

func (p *player) Run(u, title string, output io.Writer) error {
	tmpCommand := make([]string, len(p.command))
	copy(tmpCommand, p.command)

	for i := range tmpCommand {
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", u)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$title", title)
	}

	cmd := exec.Command(tmpCommand[0], tmpCommand[1:]...)

	// Override output if non nil
	if output != nil {
		cmd.Stdout = output
		cmd.Stderr = output
	}

	log.Tracef("[%s] running command [%s]", p.Name(), cmd.String())
	return cmd.Run()
}

// Each player registered in the app
// https://github.com/SoMuchForSubtlety/f1viewer/blob/master/internal/cmd/cmd.go
var players []*player

func init() {
	players = append(players, IINA.(*player))
	players = append(players, VLC.(*player))
	players = append(players, MPV.(*player))
	if runtime.GOOS == "darwin" {
		players = append(players, QuickTimePlayer.(*player))
	}
}

var (
	QuickTimePlayer Player = &player{
		name:       "QuickTime Player",
		command:    []string{"open", "-a", "quicktime player", "$url"},
		registry:   "",
		registry32: "",
	}
	IINA Player = &player{
		name:       "IINA",
		command:    []string{"iina", "--no-stdin", "$url"},
		registry:   "",
		registry32: "",
	}
	VLC Player = &player{
		name:       "VLC",
		registry:   "SOFTWARE\\VideoLAN\\VLC",
		registry32: "SOFTWARE\\WOW6432Node\\VideoLAN\\VLC",
		command:    []string{"vlc", "$url", "--meta-title=$title"},
	}
	MPV Player = &player{
		name:       "MPV",
		command:    []string{"mpv", "$url", "--quiet", "--title=$title"},
		registry:   "",
		registry32: "",
	}
)

// checkIfExist checks if player exist on $PATH or in Windows Registry
func (p *player) checkIfExist() bool {
	_, err := exec.LookPath(p.command[0])
	switch {
	case err == nil:
		return true // Found in $PATH
	case p.checkRegistry():
		return true // Found in Windows registry
	default:
		return false // Not found, cannot be used
	}
}

// DefaultPlayer return the first media player available from $PATH or Windows registry
// Throw an error when no player is available
func DefaultPlayer() (Player, error) {
	// For each player, check if found in $PATH or Windows registry and use it
	for _, player := range players {
		if ok := player.checkIfExist(); ok {
			log.Debugf("using player [%s]", player.Name())
			return player, nil
		}
	}

	return nil, fmt.Errorf("cannot find any compatible media player")
}
