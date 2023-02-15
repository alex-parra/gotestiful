package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type shArgs []string

// shCmd runs a shell command with given args and returns the output
func shCmd(prog string, args shArgs, stdIn string) (string, error) {
	cmd := exec.Command(prog, args...)

	cmd.Stdin = strings.NewReader(stdIn)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		// print out the stdout back to stdout so we can debug
		fmt.Fprintln(os.Stderr, stdErr.String())
		return "", fmt.Errorf("failed to run %s: %w", prog, err)
	}

	return stdOut.String(), nil
}

func shJSONPipe[T any](prog string, args shArgs, stdIn string, eventPipe chan<- T, copyOutput io.Writer) error {
	defer close(eventPipe)

	cmd := exec.Command(prog, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe %s: %w", prog, err)
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", prog, err)
	}

	pipedStdout := io.TeeReader(stdout, copyOutput)

	dec := json.NewDecoder(pipedStdout)
	for {
		var m T
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading shell output: %v", err)
		}

		eventPipe <- m
	}

	err = cmd.Wait()

	if err != nil {
		return fmt.Errorf("failed to run %s: %w", prog, err)
	}
	return nil
}

// shPipe runs a shell command with given args and pipes the output to a channel
func shPipe(prog string, args shArgs, stdIn string, pipeLine chan<- string) error {
	cmd := exec.Command(prog, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", prog, err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		pipeLine <- scanner.Text()
	}

	err = cmd.Wait()
	close(pipeLine)
	if err != nil {
		fmt.Fprintln(os.Stderr, stdErr.String())
		return fmt.Errorf("failed to run %s: %w", prog, err)
	}
	return nil
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
