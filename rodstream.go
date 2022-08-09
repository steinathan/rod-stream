package rodstream

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
)

func MustPrepareLauncher() *launcher.Launcher {
	var (
		extPath     = "./extension/recorder"
		extensionId = "jjndjgheafjngoipoacpjgeicjeomjli"
		absExtPath  = filepath.Join(getModPath(), extPath)
	)

	launcher := launcher.New().
		Set("allow-http-screen-capture").
		Set("enable-usermedia-screen-capturing").
		Set("whitelisted-extension-id", extensionId).
		Set("disable-extensions-except", absExtPath).
		Set("load-extension", absExtPath).
		Set("allow-google-chromefile-access").
		Headless(false)

	return launcher
}

func GrantPermissions(urls []string, browser *rod.Browser) error {
	if len(urls) == 0 {
		return errors.New("at least one url is required")
	}

	for _, url := range urls {
		_ = proto.BrowserGrantPermissions{
			Origin: url,
			Permissions: []proto.BrowserPermissionType{
				proto.BrowserPermissionTypeVideoCapture,
				proto.BrowserPermissionTypeAudioCapture,
			},
		}.Call(browser)
	}

	return nil
}

type StreamConstraints struct {
	Audio              bool   `json:"audio"`
	Video              bool   `json:"video"`
	MimeType           string `json:"mimeType,omitempty"`
	AudioBitsPerSecond int    `json:"audioBitsPerSecond,omitempty"`
	VideoBitsPerSecond int    `json:"videoBitsPerSecond,omitempty"`
	BitsPerSecond      int    `json:"bitsPerSecond,omitempty"`
	FrameSize          int    `json:"frameSize,omitempty"`
}

// MustCreatePage Must call the browser to capture the extension first handshake
// returns a page that can be used to capture video
func MustCreatePage(browser *rod.Browser) *rod.Page {
	var (
		targets, _       = proto.TargetGetTargets{}.Call(browser)
		videoCapturePage *rod.Page
	)

	for _, t := range targets.TargetInfos {
		if t.Type == proto.TargetTargetInfoTypeBackgroundPage && t.Title == "Video Streamer" {
			log.Println("Video stream extension found:", t.Title)
			videoCapturePage = browser.MustPageFromTargetID(t.TargetID)

			//goland:noinspection GoUnhandledErrorResult
			err := proto.PageBringToFront{}.Call(videoCapturePage)
			if err != nil {
				log.Panicln("Failed to focus tab!")
			}

			break
		}
	}

	return videoCapturePage
}

// MustGetStream Gets a stream from the browser's page
func MustGetStream(videoCapturePage *rod.Page, streamConstraints *StreamConstraints, ch chan string) {
	pInfo := videoCapturePage.MustInfo()
	if pInfo.Type != proto.TargetTargetInfoTypeBackgroundPage {
		log.Panicln("Page is not a background page, cannot get stream!")
	}

	if pInfo.Title != "Video Streamer" {
		log.Panicln("Page is not the video streamer, cannot get stream!")
	}

	err := proto.PageBringToFront{}.Call(videoCapturePage)
	if err != nil {
		log.Fatalln("Failed to focus tab!")
	}

	log.Println("Video stream extension created:", videoCapturePage.TargetID)

	// Start video capture
	constraints := gson.New(streamConstraints)
	js := `
	(function (pageId, contraints) {
		contraints = contraints || {
			audio: true,
			video: true,
			mimeType: "video/webm;codecs=vp8,opus",
			audioBitsPerSecond: null,
			videoBitsPerSecond: null,
			bitsPerSecond: null,
			frameSize: 500,
		}
		// Index for recording
		contraints.index = pageId
		window.START_RECORDING(contraints)
	})
	`
	videoCapturePage.MustEval(js, videoCapturePage.TargetID, constraints)

	// Get captured whole data
	videoCapturePage.MustExpose("sendWholeData", func(data gson.JSON) (interface{}, error) {
		if data.Has("type") && data.Has("chunk") {
			chunk := data.Get("chunk").String()
			ch <- chunk
		}

		return nil, nil
	})

}

// MustStopStream Stops the Stream
func MustStopStream(videoCapturePage *rod.Page, ch chan string) error {
	fmt.Println("Stopping stream", videoCapturePage.TargetID, len(ch))
	var pageId proto.TargetTargetID = videoCapturePage.TargetID
	js := `(function (pageId) {
		window.STOP_RECORDING(pageId)
	})`
	_, err := videoCapturePage.Eval(js, pageId)
	return err
}

// GetStdInWriter returns a writer that writes to stdin
func GetStdInWriter(filename string) (io.WriteCloser, error) {
	cmd := exec.Command("ffmpeg", "-y",
		// "-hide_banner",
		// "-loglevel",
		// "panic", // Hide all logs
		"-i", "pipe:0",
		filename,
	)
	videoBuffer := bytes.NewBuffer(make([]byte, 1024))

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

	return stdin, nil
}

// Parseb64 removes b64 prefix and returns decoded data
func Parseb64(data string) []byte {
	b64data := data[strings.IndexByte(data, ',')+1:]
	f, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		panic(err)
	}
	return f
}

// getModPath gets the path to the module that contains the extension
func getModPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	return filepath.Dir(filename)
}
