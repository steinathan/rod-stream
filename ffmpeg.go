package rodstream

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func GetStdInWriter(filename string) (io.WriteCloser, error) {

	cmd := exec.Command("ffmpeg", "-y",
		// "-hide_banner",
		// "-loglevel",
		// "panic", // Hide all logs
		"-i", "pipe:0",
		filename,
	)
	videoBuffer := bytes.NewBuffer(make([]byte, 5))

	// bind log stream to stderr
	cmd.Stderr = os.Stderr
	cmd.Stdout = videoBuffer

	// Open stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// Start a process on another goroutine
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// cmd.Wait()
	return stdin, nil
}
