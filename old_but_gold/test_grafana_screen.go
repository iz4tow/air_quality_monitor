package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"
	"log" // Import the log package

	"github.com/chromedp/chromedp"
)

func main() {
	// Create an allocator with the "headless" flag set to false to disable headless mode
	allocCtx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("headless", false), // Disable headless mode (show the browser)
	)
	defer cancel()

	// Create the main context with the exec allocator
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf), // Enables logging of all chromedp events
		chromedp.WithDebugf(log.Printf), // Debugging log output for browser actions
	)
	defer cancel()

	// Set a timeout for the operation
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second) // Increase timeout to 60 seconds
	defer cancel()

	// Define the URL and the output file
	url := "http://localhost:3000" // Replace with your Grafana URL
	outputFile := "screenshot.jpeg"

	// Take the screenshot
	var buf []byte
	err := chromedp.Run(ctx, takeScreenshot(url, &buf))
	if err != nil {
		fmt.Printf("Failed to take screenshot: %v\n", err)
		return
	}

	// Debug: Print the first few bytes of the buffer
	if len(buf) >= 8 {
		fmt.Printf("Screenshot data (first 8 bytes): %s\n", hex.EncodeToString(buf[:8]))
	} else {
		fmt.Println("Screenshot buffer is empty or invalid!")
		return
	}

	// Save the screenshot to a file
	err = os.WriteFile(outputFile, buf, 0644)
	if err != nil {
		fmt.Printf("Failed to save screenshot: %v\n", err)
		return
	}

	fmt.Printf("Screenshot saved to %s\n", outputFile)
}

// takeScreenshot logs in and takes a screenshot of the dashboard
func takeScreenshot(url string, buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`input[name="user"]`, chromedp.ByQuery), // Wait for the username input to appear
		chromedp.SendKeys(`input[name="user"]`, "admin", chromedp.ByQuery), // Enter username
		chromedp.SendKeys(`input[name="password"]`, "Password123!", chromedp.ByQuery), // Enter password
		chromedp.SendKeys(`input[name="password"]`, "\r", chromedp.ByQuery), // Press Enter on the password field
		chromedp.WaitVisible(`.dashboard-container`, chromedp.ByClass), // Wait for the main dashboard container to appear
		chromedp.WaitVisible(`http://localhost:3000/d/be9yswpcby39ca`, chromedp.ByQuery), // Wait for the navigation bar to ensure the page loaded
		chromedp.FullScreenshot(buf, 90), // Capture the screenshot with 90% quality
	}
}


