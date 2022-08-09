package rodstream

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func MustPrepareLauncher() *launcher.Launcher {
	var (
		extPath     = "./extension/recorder"
		extensionId = "jjndjgheafjngoipoacpjgeicjeomjli"
		absExtPath  = filepath.Join(GetModPath(), extPath)
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

// gets the path to the module that contains the extension
func GetModPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	return filepath.Dir(filename)
}
