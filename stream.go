package rodstream

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
)

type StreamConstraints struct {
	Audio              bool   `json:"audio"`
	Video              bool   `json:"video"`
	MimeType           string `json:"mimeType"`
	AudioBitsPerSecond int    `json:"audioBitsPerSecond"`
	VideoBitsPerSecond int    `json:"videoBitsPerSecond"`
	BitsPerSecond      int    `json:"bitsPerSecond"`
	FrameSize          int    `json:"frameSize"`
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
			fmt.Println("Video stream extension found:", t.Title)
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

func MustGetStream(videoCapturePage *rod.Page, streamConstraints *StreamConstraints, ch chan string) {
	var pageId proto.TargetTargetID = videoCapturePage.TargetID

	//goland:noinspection GoUnhandledErrorResult
	err := proto.PageBringToFront{}.Call(videoCapturePage)
	if err != nil {
		log.Fatalln("Failed to focus tab!")
	}

	log.Println("Video stream extension created:", videoCapturePage.TargetID)

	// Start video capture
	constraints := gson.New(streamConstraints)
	videoCapturePage.MustEval(ReadJs("./js/start-recording.js"), videoCapturePage.MustWaitIdle(), constraints)

	// Get captured whole data
	videoCapturePage.MustExpose("sendWholeData", func(data gson.JSON) (interface{}, error) {
		if data.Has("type") && data.Has("chunk") {
			chunk := data.Get("chunk").String()
			ch <- chunk
		}

		return nil, nil
	})

	fmt.Println("Video stream extension created:", pageId)
}

func MustStopStream(videoCapturePage *rod.Page) {
	var pageId proto.TargetTargetID = videoCapturePage.TargetID
	videoCapturePage.MustEval(ReadJs("./js/stop-recording.js"), pageId)
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

func ReadJs(path string) string {
	p := filepath.Join(GetModPath(), path)
	js, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalln(err)
	}
	return string(js)
}
