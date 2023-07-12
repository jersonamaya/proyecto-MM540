package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"

	"gocv.io/x/gocv"
)

const (
	imagesPath      = "C:/Users/Gaby/Desktop/Extrayendo rostros/Imagenes"
	facesFolderPath = "Rostros encontrados"
)

func main() {
	if _, err := os.Stat(facesFolderPath); os.IsNotExist(err) {
		fmt.Println("Carpeta creada:", facesFolderPath)
		os.Mkdir(facesFolderPath, 0755)
	}

	faceCascade := gocv.NewCascadeClassifier()
	defer faceCascade.Close()

	if !faceCascade.Load(gocv.LoadConfig{
		ClassifierFile: filepath.Join(gocv.OpenCVHaarCascadeData, "haarcascade_frontalface_default.xml"),
	}) {
		fmt.Println("Error al cargar el clasificador de detección de rostros")
		return
	}

	imagesPathList, err := filepath.Glob(filepath.Join(imagesPath, "*.jpg"))
	if err != nil {
		fmt.Println("Error al obtener la lista de imágenes:", err)
		return
	}

	window := gocv.NewWindow("image")
	defer window.Close()

	count := 0
	for _, imagePath := range imagesPathList {
		imageMat := gocv.IMRead(imagePath, gocv.IMReadColor)
		if imageMat.Empty() {
			fmt.Println("No se puede leer la imagen:", imagePath)
			continue
		}

		gray := gocv.NewMat()
		defer gray.Close()
		gocv.CvtColor(imageMat, &gray, gocv.ColorBGRToGray)

		faces := faceCascade.DetectMultiScaleWithParams(gray, 1.1, 5, 0, image.Pt(0, 0), image.Pt(0, 0))

		for _, r := range faces {
			gocv.Rectangle(&imageMat, r, color.RGBA{R: 128, G: 0, B: 255, A: 0}, 2)
		}

		gocv.Rectangle(&imageMat, image.Rect(10, 5, 450, 25), color.RGBA{R: 255, G: 255, B: 255, A: 0}, -1)
		gocv.PutText(&imageMat, "Presione s, para alamacenar los rostros encontrados", image.Pt(10, 20), gocv.FontHersheySimplex, 0.5, color.RGBA{R: 128, G: 0, B: 255, A: 0}, 1)

		window.IMShow(imageMat)
		key := window.WaitKey(0)

		if key == 's' {
			for i, r := range faces {
				rostro := imageMat.Region(r)
				resizedRostro := gocv.NewMat()
				defer resizedRostro.Close()
				gocv.Resize(rostro, &resizedRostro, image.Point{X: 150, Y: 150}, 0, 0, gocv.InterpolationCubic)
				gocv.IMWrite(filepath.Join(facesFolderPath, fmt.Sprintf("rostro_%d.jpg", count+i)), resizedRostro)
			}
			count += len(faces)
		} else if key == 27 {
			break
		}
	}

	window.Close()
}
