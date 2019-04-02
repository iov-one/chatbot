package chatbot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var splitRegex = regexp.MustCompile("/s+")

func isNotFound(str string) bool {
	return strings.Contains(str, "NotFound")
}

func Log(str string, args ...interface{}) {
	logStr := fmt.Sprintf("[%s] %s\n", time.Now(), str)
	log.Printf(logStr, args...)
}

func execute(commandStr string) string {
	var commandArgs []string
	commandSlice := strings.Fields(commandStr)
	command := commandSlice[0]

	if len(commandSlice) > 1 {
		commandArgs = commandSlice[1:]
	}

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	cmd := exec.Command(command, commandArgs...)

	stdin, err := cmd.StdinPipe()

	if err != nil {
		Log(err.Error())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		Log("error executing command: %s", err.Error())
	}

	stdin.Close()
	cmd.Wait()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out)
}
