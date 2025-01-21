package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// Create a context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set a timeout for the operation
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Define the URL and the output file
	url := "https://example.com"
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

// takeScreenshot takes a screenshot of the specified URL
func takeScreenshot(url string, buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(2 * time.Second), // Wait for the page to load
		chromedp.FullScreenshot(buf, 90), // Capture the screenshot with 90% quality
	}
}

