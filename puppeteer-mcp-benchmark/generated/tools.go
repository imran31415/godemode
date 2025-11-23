package browsertools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserContext holds the browser session
type BrowserContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var browserCtx *BrowserContext

// InitBrowser initializes a new browser context
func InitBrowser() error {
	if browserCtx != nil {
		return nil // Already initialized
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	browserCtx = &BrowserContext{
		ctx:    ctx,
		cancel: cancel,
	}

	return nil
}

// CloseBrowser closes the browser
func CloseBrowser() {
	if browserCtx != nil {
		browserCtx.cancel()
		browserCtx = nil
	}
}

// Navigate goes to a URL and waits for the page to load
func Navigate(url string, waitSelector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		if err := InitBrowser(); err != nil {
			return nil, err
		}
	}

	var actions []chromedp.Action
	actions = append(actions, chromedp.Navigate(url))

	if waitSelector != "" {
		actions = append(actions, chromedp.WaitVisible(waitSelector, chromedp.ByQuery))
	} else {
		actions = append(actions, chromedp.WaitReady("body", chromedp.ByQuery))
	}

	err := chromedp.Run(browserCtx.ctx, actions...)
	if err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	var currentURL string
	chromedp.Run(browserCtx.ctx, chromedp.Location(&currentURL))

	return map[string]interface{}{
		"success": true,
		"url":     currentURL,
	}, nil
}

// GetTitle gets the page title
func GetTitle() (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var title string
	err := chromedp.Run(browserCtx.ctx, chromedp.Title(&title))
	if err != nil {
		return nil, fmt.Errorf("failed to get title: %w", err)
	}

	return map[string]interface{}{
		"title": title,
	}, nil
}

// GetURL gets the current URL
func GetURL() (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var url string
	err := chromedp.Run(browserCtx.ctx, chromedp.Location(&url))
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return map[string]interface{}{
		"url": url,
	}, nil
}

// Screenshot takes a screenshot of the page
func Screenshot(filename string, fullPage bool) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var buf []byte
	var err error

	if fullPage {
		err = chromedp.Run(browserCtx.ctx, chromedp.FullScreenshot(&buf, 90))
	} else {
		err = chromedp.Run(browserCtx.ctx, chromedp.CaptureScreenshot(&buf))
	}

	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	// Save to file
	if filename != "" {
		dir := filepath.Dir(filename)
		os.MkdirAll(dir, 0755)
		err = os.WriteFile(filename, buf, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to save screenshot: %w", err)
		}
	}

	return map[string]interface{}{
		"success":  true,
		"filename": filename,
		"size":     len(buf),
	}, nil
}

// GetText extracts text from an element
func GetText(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var text string
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get text: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"text":     text,
	}, nil
}

// GetAttribute gets an attribute value from an element
func GetAttribute(selector string, attribute string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var value string
	var ok bool
	err := chromedp.Run(browserCtx.ctx,
		chromedp.AttributeValue(selector, attribute, &value, &ok, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	return map[string]interface{}{
		"selector":  selector,
		"attribute": attribute,
		"value":     value,
		"exists":    ok,
	}, nil
}

// Click clicks on an element
func Click(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	err := chromedp.Run(browserCtx.ctx,
		chromedp.Click(selector, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("click failed: %w", err)
	}

	return map[string]interface{}{
		"success":  true,
		"selector": selector,
	}, nil
}

// Type types text into an input field
func Type(selector string, text string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	err := chromedp.Run(browserCtx.ctx,
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("type failed: %w", err)
	}

	return map[string]interface{}{
		"success":  true,
		"selector": selector,
		"text":     text,
	}, nil
}

// WaitForSelector waits for an element to appear
func WaitForSelector(selector string, timeout int) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("wait failed: %w", err)
	}

	return map[string]interface{}{
		"success":  true,
		"selector": selector,
	}, nil
}

// CountElements counts elements matching a selector
func CountElements(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var count int
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelectorAll('%s').length`, selector), &count),
	)
	if err != nil {
		return nil, fmt.Errorf("count failed: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"count":    count,
	}, nil
}

// GetAllText extracts text from all elements matching a selector
func GetAllText(selector string, limit int) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var texts []string
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			Array.from(document.querySelectorAll('%s'))
				.slice(0, %d)
				.map(el => el.innerText.trim())
		`, selector, limit), &texts),
	)
	if err != nil {
		return nil, fmt.Errorf("get all text failed: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"texts":    texts,
		"count":    len(texts),
	}, nil
}

// GetAllAttributes extracts an attribute from all elements matching a selector
func GetAllAttributes(selector string, attribute string, limit int) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var values []string
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			Array.from(document.querySelectorAll('%s'))
				.slice(0, %d)
				.map(el => el.getAttribute('%s') || '')
		`, selector, limit, attribute), &values),
	)
	if err != nil {
		return nil, fmt.Errorf("get all attributes failed: %w", err)
	}

	return map[string]interface{}{
		"selector":  selector,
		"attribute": attribute,
		"values":    values,
		"count":     len(values),
	}, nil
}

// Evaluate runs JavaScript in the page context
func Evaluate(script string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var result interface{}
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Evaluate(script, &result),
	)
	if err != nil {
		return nil, fmt.Errorf("evaluate failed: %w", err)
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}

// ScrollTo scrolls to a position or element
func ScrollTo(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	err := chromedp.Run(browserCtx.ctx,
		chromedp.ScrollIntoView(selector, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("scroll failed: %w", err)
	}

	return map[string]interface{}{
		"success":  true,
		"selector": selector,
	}, nil
}

// ElementExists checks if an element exists
func ElementExists(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var exists bool
	err := chromedp.Run(browserCtx.ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, selector), &exists),
	)
	if err != nil {
		return nil, fmt.Errorf("element exists check failed: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"exists":   exists,
	}, nil
}

// GetInnerHTML gets the innerHTML of an element
func GetInnerHTML(selector string) (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var html string
	err := chromedp.Run(browserCtx.ctx,
		chromedp.InnerHTML(selector, &html, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("get innerHTML failed: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"html":     html,
	}, nil
}

// Sleep waits for a specified duration
func Sleep(milliseconds int) (map[string]interface{}, error) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	return map[string]interface{}{
		"success": true,
		"slept":   milliseconds,
	}, nil
}

// GetPageSource gets the full page HTML source
func GetPageSource() (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var html string
	err := chromedp.Run(browserCtx.ctx,
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("get page source failed: %w", err)
	}

	return map[string]interface{}{
		"length": len(html),
		"source": html[:min(500, len(html))] + "...", // Truncate for display
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Base64Screenshot returns screenshot as base64
func Base64Screenshot() (map[string]interface{}, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	var buf []byte
	err := chromedp.Run(browserCtx.ctx, chromedp.CaptureScreenshot(&buf))
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"base64":  base64.StdEncoding.EncodeToString(buf),
		"size":    len(buf),
	}, nil
}
