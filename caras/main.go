package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"image"
	"image/draw"
	"image/jpeg"

	"github.com/Kagami/go-face"
	"github.com/nfnt/resize"
)

const dataDir = "testdata"
const facesDir = "faces"                  // Directorio para guardar las caras encontradas
const facesDetectedDir = "faces_detected" // Directorio para guardar las caras detectadas

var modelsDir = filepath.Join(dataDir, "models")
var imagesDir = filepath.Join(dataDir, "images")

// Declarar el mapa para relacionar las caras detectadas con las imágenes originales.
var faceToOriginalImageMap = make(map[string]string)

func main() {
	// Inicializar el reconocedor.
	rec, err := face.NewRecognizer(modelsDir)
	if err != nil {
		log.Fatalf("No se puede inicializar el reconocedor facial: %v", err)
	}
	// Liberar los recursos cuando hayas terminado.
	defer rec.Close()

	// Manejar la ruta de carga.
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			file, _, err := r.FormFile("image")
			if err != nil {
				http.Error(w, "Error al cargar la imagen", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Guardar la imagen subida en un archivo temporal.
			tempFile, err := os.CreateTemp("", "uploaded_image_*.jpg")
			if err != nil {
				http.Error(w, "Error al guardar la imagen", http.StatusInternalServerError)
				return
			}
			defer tempFile.Close()

			_, err = io.Copy(tempFile, file)
			if err != nil {
				http.Error(w, "Error al guardar la imagen", http.StatusInternalServerError)
				return
			}

			// Reconocer caras en la imagen subida.
			faces, err := rec.RecognizeFile(tempFile.Name())
			if err != nil {
				http.Error(w, "Error al reconocer las caras", http.StatusInternalServerError)
				return
			}

			// Crear el directorio para las caras detectadas si no existe.
			err = os.MkdirAll(facesDetectedDir, 0755)
			if err != nil {
				http.Error(w, "Error al crear el directorio para las caras detectadas", http.StatusInternalServerError)
				return
			}

			// Obtener el nombre de la imagen original sin la ruta.
			originalImageName := filepath.Base(tempFile.Name())

			// Guardar las caras encontradas en archivos individuales y almacenar la relación con las imágenes originales.
			for i, face := range faces {
				faceImageFile := filepath.Join(facesDetectedDir, fmt.Sprintf("%s_face_%d.jpg", originalImageName, i+1))
				saveFaceImage(tempFile.Name(), face.Rectangle, faceImageFile)

				// Guardar la relación entre la cara detectada y la imagen original en el mapa.
				faceToOriginalImageMap[filepath.Base(faceImageFile)] = originalImageName
			}

			// Redireccionar a la página de resultados.
			http.Redirect(w, r, "/result", http.StatusSeeOther)

		} else {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	// Servir el formulario front-end.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Servir la página de resultados.
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		fileInfos, err := os.ReadDir(facesDetectedDir)
		if err != nil {
			http.Error(w, "Error al leer las imágenes de caras detectadas", http.StatusInternalServerError)
			return
		}

		var carasDetectadas []string
		for _, fileInfo := range fileInfos {
			if strings.HasSuffix(fileInfo.Name(), ".jpg") {
				carasDetectadas = append(carasDetectadas, fileInfo.Name())
			}
		}

		// Renderizar el resultado y pasar el mapa con la relación de caras detectadas e imágenes originales.
		tpl, err := template.New("result").Parse(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<title>Resultados de reconocimiento facial</title>
			</head>
			<body>
				<h1>Resultados de reconocimiento facial</h1>
				{{range $index, $label := .}}
				<img src="/faces_detected/{{.}}" alt="{{$label}}">
				<label>{{$label}}</label>
				<br>
				<!-- Mostrar el botón para ver la foto original -->
				<a href="/original_image?face={{.}}">Ver foto original</a>
				<br>
				{{end}}
			</body>
			</html>
		`)
		if err != nil {
			http.Error(w, "Error al renderizar el resultado", http.StatusInternalServerError)
			return
		}

		tpl.Execute(w, carasDetectadas)
	})

	// Servir las caras detectadas.
	http.Handle("/faces_detected/", http.StripPrefix("/faces_detected/", http.FileServer(http.Dir(facesDetectedDir))))

	// Agregar un manejador para servir las imágenes originales.
	http.HandleFunc("/original_image", func(w http.ResponseWriter, r *http.Request) {
		faceImage := r.URL.Query().Get("face")
		if originalImage, ok := faceToOriginalImageMap[faceImage]; ok {
			http.ServeFile(w, r, filepath.Join(imagesDir, originalImage))
		} else {
			http.Error(w, "Imagen no encontrada", http.StatusNotFound)
		}
	})

	// Iniciar el servidor en el puerto 8080.
	log.Println("El servidor está ejecutándose en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Función para guardar la imagen de la cara detectada.
func saveFaceImage(originalImagePath string, rect image.Rectangle, filename string) error {
	img, err := os.Open(originalImagePath)
	if err != nil {
		return err
	}
	defer img.Close()

	// Decodificar la imagen original.
	originalImage, _, err := image.Decode(img)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Convertir la imagen decodificada a *image.RGBA.
	rgbaImage := image.NewRGBA(rect)
	draw.Draw(rgbaImage, rgbaImage.Bounds(), originalImage, rect.Min, draw.Src)

	// Redimensionar la región de la cara de la imagen original a un tamaño fijo.
	const fixedWidth, fixedHeight = 200, 200
	resizedImage := resize.Resize(fixedWidth, fixedHeight, rgbaImage, resize.Lanczos3)

	// Guardar la imagen recortada y redimensionada en formato JPEG.
	return jpeg.Encode(outFile, resizedImage, nil)
}
