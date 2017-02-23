package shell

import (
	"bufio"
	"fmt"
	"os"

	"github.com/codeskyblue/go-sh"
)

type Shell struct {
	Session *sh.Session
}

func New() *Shell {
	c := &Shell{}
	c.Session = sh.NewSession()
	c.Session.ShowCMD = false
	return c
}

func (s *Shell) WinCmd(str string) string {
	cmd := s.Session.Command("cmd", "/C", str)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (s *Shell) LinuxCmd(str string) string {
	cmd := s.Session.Command("bash", "-c", str)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func Pause() {
	fmt.Println("Press Enter To Continue ...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
