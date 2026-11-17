package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
)

type ComponentSpec struct {
	Filename string
	Title    string
	Subtitle string
	Type     string
	TDP      string
	BgColor  color.RGBA
	AccColor color.RGBA
}

func drawRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

func drawFrame(img *image.RGBA, x1, y1, x2, y2, border int, bg, stroke color.Color) {
	drawRect(img, x1, y1, x2, y2, stroke)
	drawRect(img, x1+border, y1+border, x2-border, y2-border, bg)
}

func drawGrid(img *image.RGBA, step int, c color.Color) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	for x := 0; x < w; x += step {
		for y := 0; y < h; y++ {
			img.Set(x, y, c)
		}
	}
	for y := 0; y < h; y += step {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
}

func generateImage(spec ComponentSpec, outPath string) error {
	w, h := 600, 600
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	draw.Draw(img, img.Bounds(), &image.Uniform{spec.BgColor}, image.Point{}, draw.Src)

	gridCol := color.RGBA{R: spec.AccColor.R / 3, G: spec.AccColor.G / 3, B: spec.AccColor.B / 3, A: 255}
	drawGrid(img, 40, gridCol)

	switch spec.Type {
	case "CPU":
		
		drawFrame(img, 100, 100, 500, 500, 12, color.RGBA{45, 52, 60, 255}, spec.AccColor)
		drawFrame(img, 140, 140, 460, 460, 6, color.RGBA{30, 36, 44, 255}, color.RGBA{180, 150, 80, 255})
		
		drawRect(img, 110, 110, 130, 130, spec.AccColor)
		
		drawFrame(img, 180, 180, 420, 420, 4, color.RGBA{22, 27, 34, 255}, spec.AccColor)
		
		drawRect(img, 210, 210, 390, 390, color.RGBA{15, 20, 28, 255})

		for i := 115; i < 485; i += 16 {
			drawRect(img, i, 103, i+8, 110, spec.AccColor)
			drawRect(img, i, 490, i+8, 497, spec.AccColor)
			drawRect(img, 103, i, 110, i+8, spec.AccColor)
			drawRect(img, 490, i, 497, i+8, spec.AccColor)
		}

	case "GPU":
		drawFrame(img, 80, 140, 520, 460, 8, color.RGBA{35, 42, 38, 255}, spec.AccColor)
		
		drawRect(img, 120, 460, 480, 485, color.RGBA{220, 175, 45, 255})
		for i := 125; i < 475; i += 8 {
			drawRect(img, i, 485, i+4, 495, color.RGBA{240, 200, 70, 255})
		}
		for i := 100; i < 500; i += 14 {
			drawRect(img, i, 160, i+6, 440, color.RGBA{70, 85, 75, 255})
		}
		drawFrame(img, 220, 240, 380, 360, 4, color.RGBA{20, 30, 24, 255}, spec.AccColor)

	case "PSU":
		drawFrame(img, 90, 160, 510, 440, 8, color.RGBA{40, 40, 46, 255}, spec.AccColor)
		
		drawFrame(img, 130, 200, 330, 400, 6, color.RGBA{20, 20, 26, 255}, spec.AccColor)

		drawRect(img, 215, 210, 245, 390, color.RGBA{60, 60, 70, 255})
		drawRect(img, 140, 285, 320, 315, color.RGBA{60, 60, 70, 255})
		
		drawRect(img, 360, 200, 480, 240, color.RGBA{200, 50, 50, 255})
		
		drawFrame(img, 360, 270, 470, 350, 4, color.RGBA{15, 15, 20, 255}, spec.AccColor)
		
		drawRect(img, 370, 380, 390, 400, color.RGBA{50, 220, 90, 255})
		drawRect(img, 410, 380, 430, 400, color.RGBA{230, 180, 30, 255})

	default
		drawFrame(img, 120, 100, 480, 500, 8, color.RGBA{38, 34, 32, 255}, spec.AccColor)
		for y := 130; y < 470; y += 70 {
			drawFrame(img, 140, y, 460, y+55, 4, color.RGBA{24, 22, 20, 255}, spec.AccColor)
			for x := 155; x < 350; x += 40 {
				drawRect(img, x, y+15, x+28, y+40, color.RGBA{55, 50, 48, 255})
			}
			drawRect(img, 420, y+20, 435, y+35, color.RGBA{80, 210, 100, 255})
		}
	}

	drawFrame(img, 20, 20, 580, 85, 3, color.RGBA{18, 16, 15, 230}, spec.AccColor)

	drawFrame(img, 20, 515, 580, 580, 3, color.RGBA{18, 16, 15, 230}, spec.AccColor)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func main() {
	outDir := filepath.Join("resources", "img")
	_ = os.MkdirAll(outDir, 0755)

	specs := []ComponentSpec{
		{
			Filename: "cpu-xeon.jpg",
			Title:    "Intel Xeon E5-2680 v4",
			Subtitle: "14 Cores / 28 Threads",
			Type:     "CPU",
			TDP:      "120 W",
			BgColor:  color.RGBA{18, 28, 45, 255},  
			AccColor: color.RGBA{0, 150, 255, 255}, 
		},
		{
			Filename: "gpu-tesla.jpg",
			Title:    "NVIDIA Tesla V100",
			Subtitle: "Volta Architecture / 32GB HBM2",
			Type:     "GPU",
			TDP:      "250 W",
			BgColor:  color.RGBA{20, 36, 24, 255},  
			AccColor: color.RGBA{118, 185, 0, 255}, 
		},
		{
			Filename: "psu-800.jpg",
			Title:    "Power Supply 800W",
			Subtitle: "80 PLUS Platinum Hot-Swap",
			Type:     "PSU",
			TDP:      "800 W",
			BgColor:  color.RGBA{38, 30, 20, 255},  
			AccColor: color.RGBA{235, 175, 40, 255}, 
		},
		{
			Filename: "cpu-epyc.jpg",
			Title:    "AMD EPYC 7742",
			Subtitle: "64 Cores / 128 Threads",
			Type:     "CPU",
			TDP:      "225 W",
			BgColor:  color.RGBA{45, 22, 18, 255},  
			AccColor: color.RGBA{237, 56, 36, 255},  
		},
		{
			Filename: "gpu-rtx.jpg",
			Title:    "NVIDIA RTX A6000",
			Subtitle: "Ampere Architecture / 48GB",
			Type:     "GPU",
			TDP:      "300 W",
			BgColor:  color.RGBA{28, 24, 38, 255},  
			AccColor: color.RGBA{170, 100, 245, 255},
		},
		{
			Filename: "default.jpg",
			Title:    "Server Cooling System",
			Subtitle: "Default SSR Media / Fallback",
			Type:     "DEFAULT",
			TDP:      "N/A",
			BgColor:  color.RGBA{43, 37, 34, 255},  
			AccColor: color.RGBA{200, 157, 92, 255}, 
		},
	}

	for _, spec := range specs {
		path := filepath.Join(outDir, spec.Filename)
		err := generateImage(spec, path)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", spec.Filename, err)
		} else {
			fmt.Printf("Generated %s successfully!\n", spec.Filename)
		}
	}

	defaultVid := filepath.Join(outDir, "default.mp4")
	if _, err := os.Stat(defaultVid); err == nil {
		videoFiles := []string{
			"cpu-xeon.mp4",
			"gpu-tesla.mp4",
			"psu-800.mp4",
			"cpu-epyc.mp4",
			"gpu-rtx.mp4",
		}
		for _, vf := range videoFiles {
			target := filepath.Join(outDir, vf)
			_ = copyFile(defaultVid, target)
			fmt.Printf("Prepared video file %s\n", vf)
		}
	}

	fmt.Println("All media files generated successfully!")
}
