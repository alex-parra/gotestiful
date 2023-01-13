package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type shArgs []string

// shCmd runs a shell command with given args and returns the output
func shCmd(prog string, args shArgs, stdIn string) string {
	cmd := exec.Command(prog, args...)

	cmd.Stdin = strings.NewReader(stdIn)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to run %s: %w", prog, err))
	}

	return stdOut.String()
}

// shPipe runs a shell command with given args and pipes the output to a channel
func shPipe(prog string, args shArgs, stdIn string, pipeLine chan string) {
	cmd := exec.Command(prog, args...)

	stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to run %s: %w", prog, err))
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		pipeLine <- scanner.Text()
	}

	err = cmd.Wait()
	close(pipeLine)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to run %s: %w", prog, err))
	}
}

// shColor adds ansi escape codes to colorize shell output
func shColor(fx, str string, a ...any) string {
	if color.NoColor {
		return str
	}

	opts := strings.Split(fx, ":")
	colorName := sliceAt(opts, 0, "reset")
	effect := sliceAt(opts, 1, "")

	whiteSmoke := func(s string, a ...any) string {
		if len(a) > 0 {
			s = sf(s, a...)
		}
		return "\033[38;2;180;180;180m" + s + "\033[39m"
	}

	gray := func(s string, a ...any) string {
		if len(a) > 0 {
			s = sf(s, a...)
		}
		return "\033[38;2;85;85;85m" + s + "\033[39m"
	}

	colors := map[string]func(s string, a ...any) string{
		"red":        color.RedString,
		"green":      color.GreenString,
		"yellow":     color.YellowString,
		"blue":       color.BlueString,
		"purple":     color.MagentaString,
		"cyan":       color.CyanString,
		"white":      color.WhiteString,
		"whitesmoke": whiteSmoke,
		"gray":       gray,
		"reset":      sf,
	}

	if effect == "bold" {
		str = "\033[1m" + str + "\033[22m"
	}

	if _, ok := colors[colorName]; !ok {
		log.Printf("WARN: unsupported color '%s'", colorName)
		colorName = "reset"
	}

	return colors[colorName](str, a...)
}
