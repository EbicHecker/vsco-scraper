package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"os"
	"sync/atomic"
	"time"
)

var (
	username = os.Args[1]
	browser  = rod.New().MustConnect()

	_ = os.Mkdir(username, os.ModeDir)
)

func main() {
	page := browser.MustPage(fmt.Sprintf("https://vsco.co/%s/gallery", username)).MustWaitLoad().MustSetCookies(&proto.NetworkCookieParam{
		Name:   "OptanonAlertBoxClosed",
		Value:  fmt.Sprintf("%s.874Z", time.Now().Format("2006-01-02T15:04:05")),
		Domain: "vsco.co",
		Path:   "/",
	}, &proto.NetworkCookieParam{
		Name:   "OptanonConsent",
		Value:  "isGpcEnabled=0&datestamp=&version=6.36.0&isIABGlobal=false&hosts=&consentId=&interactionCount=1&AwaitingReconsent=false",
		Domain: "vsco.co",
		Path:   "/",
	}).MustReload()
	page.MustElement(".navbar").MustRemove()
	defer page.Close()

	var (
		index, retries int32
		count          int
	)

	cache := make([]string, 0)

	for {
		elements := page.MustWaitLoad().MustElements(".MediaImage")
		if len(elements) < 1 || index > 1 && count == len(elements) {
			if atomic.AddInt32(&retries, 1) > 10 {
				return
			}
			time.Sleep(time.Second)
			continue
		}
		count = len(elements)

	SCRAPER:
		for _, element := range elements {
			html := element.MustWaitLoad().MustHTML()
			for _, cached := range cache {
				if cached == html {
					continue SCRAPER
				}
			}
			element.MustScrollIntoView().MustScreenshot(fmt.Sprintf("%s\\image_%v.png", username, atomic.AddInt32(&index, 1)))
			cache = append(cache, html)
		}

		timer := time.NewTimer(500 * time.Millisecond)
		ticker := time.NewTicker(100 * time.Millisecond)

	MORE:
		for {
			select {
			case <-timer.C:
				break MORE
			case <-ticker.C:
				if exists, button, _ := page.Has(".css-1qxd92v"); exists {
					button.MustClick()
					break MORE
				}
			}
		}
	}
}
