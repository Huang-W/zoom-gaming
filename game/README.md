### Game Server

#### Instructions

- [Install latest version of Go](https://golang.org/doc/install)
- `go env -w GO111MODULE=on`
- `go run main.go server.go`
- `echo KERNEL==\"uinput\", GROUP=\"$USER\", MODE:=\"0660\" | sudo tee /etc/udev/rules.d/99-$USER.rules` Linux uinput
- `sudo udevadm trigger`

#### Keyboard Mappings for demo game

- [PCGamingWiki - Lovers in a Dangerous Spacetime](https://www.pcgamingwiki.com/wiki/Lovers_in_a_Dangerous_Spacetime)

#### Tests

- `go test -v zoomgaming/utils`
- `go test -v -race zoomgaming/websocket`
