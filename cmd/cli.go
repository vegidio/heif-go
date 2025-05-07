package main

import (
	"fmt"
	"github.com/vegidio/heif-go"
	"image"
	"os"
	"time"
)

func main() {
	// Encode a JPEG image to HEIC format
	encode("assets/image.jpg", "assets/image.heic")
}

func encode(inputFile, outputFile string) {
	jpgFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("failed to open JPEG -", err)
		return
	}

	defer jpgFile.Close()

	jpgImg, _, err := image.Decode(jpgFile)
	if err != nil {
		fmt.Println("failed to decode JPEG -", err)
		return
	}

	heicFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("failed to create HEIF -", err)
		return
	}

	defer heicFile.Close()

	start := time.Now()

	err = heif.Encode(heicFile, jpgImg, nil)
	if err != nil {
		fmt.Println("failed to encode HEIF -", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("Encoding completed in %s\n", duration)
}
