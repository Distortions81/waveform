package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
)

// Function to calculate the intensity at a point on the screen due to two wave sources
func calculateInterference(y, slit1Y, slit2Y, w1, w2, distance float64) float64 {
	// Calculate distances from the screen point (y) to the two slits (both positioned along the y-axis)
	distance1 := math.Sqrt(math.Pow(slit1Y-y, 2) + math.Pow(distance, 2))
	distance2 := math.Sqrt(math.Pow(slit2Y-y, 2) + math.Pow(distance, 2))

	// Calculate the phase differences for both slits
	phase1 := 2 * math.Pi * distance1 / w1
	phase2 := 2 * math.Pi * distance2 / w2

	// Amplitude is the sum of the two wave contributions
	amplitude := math.Sin(phase1) + math.Sin(phase2)

	// The intensity is the square of the amplitude
	intensity := amplitude * amplitude

	return intensity
}

func main() {
	const width, height = 216, 3840
	const center = (height / 2)
	const offset = (height / 4)                             // Screen dimensions (narrow, to mimic panel c)
	const slit1Y, slit2Y = center + offset, center - offset // Slit positions
	const distance = 300.0                                  // Distance from the slits to the screen (detector)
	const numFrames = 6000
	const freqDiv = 100.0

	os.Mkdir("render", 0755)
	os.Remove("output.mp4")

	for x := 1; x < numFrames; x++ {
		// Create a new image to represent the screen
		img := image.NewRGBA(image.Rect(0, 0, height, width)) // Note: Swap width and height

		// Iterate over each pixel along the height (y-axis) to simulate the intensity on the screen
		for y := 0; y < height; y++ {
			// Calculate the interference intensity at this point on the screen
			intensity := calculateInterference(float64(y), slit1Y, slit2Y, float64(x)/freqDiv, float64(x)/freqDiv, distance)

			// Normalize the intensity to a value between 0 and 255 for grayscale rendering
			grayValue := uint8(math.Min(intensity*255/4, 255))    // Scaling factor for visibility
			color := color.RGBA{R: grayValue, G: 0, B: 0, A: 255} // Red channel for bright spots

			// Set the pixel color for the rotated image (swap x and y to rotate 90 degrees CCW)
			for x := 0; x < width; x++ {
				img.Set(y, width-x-1, color) // Swap and reverse the x axis for 90 degree CCW rotation
			}
		}

		// Save the generated image as PNG
		fileName := fmt.Sprintf("render/frame_%03d.png", x)
		file, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Encode the image to PNG format and save it
		png.Encode(file, img)
		fmt.Println("Image saved:", fileName)
	}
	compressImagesToVideo("render/frame_%03d.png", "output.mp4", 60, 12)
}

func compressImagesToVideo(inputPattern string, outputFile string, frameRate int, crf int) error {
	// Build the FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-framerate", fmt.Sprintf("%d", frameRate),
		"-i", inputPattern,
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-pix_fmt", "yuv420p",
		outputFile,
	)

	// Set up the standard output and error streams
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running FFmpeg command: %v", err)
	}

	return nil
}
