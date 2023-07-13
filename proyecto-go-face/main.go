package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Kagami/go-face"
)

const (
	imagesPath      = "C:/Users/amaya/go/src/proyecto1"
	facesFolderPath = "Rostros encontrados"
)

type DetectionResult struct {
	Count int `json:"count"`
}

func main() {
	if _, err := os.Stat(facesFolderPath); os.IsNotExist(err) {
		fmt.Println("Carpeta creada:", facesFolderPath)
		os.Mkdir(facesFolderPath, 0755)
	}

	rec, err := face.NewRecognizer()
	if err != nil {
		fmt.Println("Error al crear el reconocedor de rostros")
		return
	}
	defer rec.Close()

	modelFile := filepath.Join("model.xml")
	err = rec.Load(modelFile)
	if err != nil {
		fmt.Println("Error al cargar el modelo de reconocimiento facial")
		return
	}

	http.HandleFunc("/detect", func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("image")
		if err != nil {
		  http.Error(w, "Error al obtener el archivo", http.StatusBadRequest)
		  return
		}
		defer file.Close()
	  
		image, _, err := image.Decode(file)
		if err != nil {
		  http.Error(w, "Error al leer la imagen", http.StatusBadRequest)
		  return
		}
	  
		// Resto del código de detección y reconocimiento facial...
	  
		// Respuesta JSON con los resultados
		response := DetectionResult{
		  Count: count,
		}
	  
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	  })
	  

	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Servidor escuchando en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func drawRectangle(img image.Image, rect image.Rectangle, col color.RGBA) {
	for x := rect.Min.X; x <= rect.Max.X; x++ {
		img.Set(x, rect.Min.Y, col)
		img.Set(x, rect.Max.Y, col)
	}
	for y := rect.Min.Y; y <= rect.Max.Y; y++ {
		img.Set(rect.Min.X, y, col)
		img.Set(rect.Max.X, y, col)
	}
}
