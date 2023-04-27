package kiosk

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// NewRelicKiosk creates a chrome-based kiosk using a local New Relic account.
func NewRelicKiosk(cfg *Config, messages chan string) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	log.Println("Using temp dir:", dir)
	defer os.RemoveAll(dir)

	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	log.Println("Navigating to ", cfg.Target.URL)
	/*
		Launch chrome and login with local user account

		name=user, type=text
		id=inputPassword, type=password, name=password
	*/
	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(cfg.Target.PreURL),
		chromedp.WaitVisible(`//input[@id="login_email"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@id="login_email"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.Click(`//*[@id="login_submit"]`, chromedp.BySearch),
	); err != nil {
		panic(err)
	}

	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@id="login_password"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@id="login_password"]`, cfg.Target.Password, chromedp.BySearch),
		chromedp.Click(`//*[@id="login_submit"]`, chromedp.BySearch),
	); err != nil {
		panic(err)
	}

	time.Sleep(10000 * time.Millisecond)

	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//button[aria-label="User menu"]`, chromedp.BySearch),
		chromedp.Click(`//button[aria-label="User menu"]`, chromedp.BySearch),
	); err != nil {
		panic(err)
	}

	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//button[aria-label="Dark"]`, chromedp.BySearch),
		chromedp.Click(`//button[aria-label="Dark"]`, chromedp.BySearch),
	); err != nil {
		panic(err)
	}

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(cfg.Target.URL),
	); err != nil {
		panic(err)
	}

	// blocking wait
	for {
		messageFromChrome := <-messages
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(cfg.Target.URL),
		); err != nil {
			panic(err)
		}
		log.Println("Chromium output:", messageFromChrome)
	}
}
