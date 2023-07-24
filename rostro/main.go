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
)

const dataDir = "testdata"
const facesDir = "faces" // Directorio para guardar las caras encontradas

var modelsDir = filepath.Join(dataDir, "models")
var imagesDir = filepath.Join(dataDir, "images")

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

			// Crear el directorio para las caras si no existe.
			err = os.MkdirAll(facesDir, 0755)
			if err != nil {
				http.Error(w, "Error al crear el directorio para las caras", http.StatusInternalServerError)
				return
			}

			// Abrir la imagen original para recortar las caras.
			img, err := os.Open(tempFile.Name())
			if err != nil {
				http.Error(w, "Error al abrir la imagen original", http.StatusInternalServerError)
				return
			}
			defer img.Close()

			// Decodificar la imagen original.
			originalImage, _, err := image.Decode(img)
			if err != nil {
				http.Error(w, "Error al decodificar la imagen original", http.StatusInternalServerError)
				return
			}

			// Guardar las caras encontradas en archivos individuales.
			for i, face := range faces {
				faceImageFile := filepath.Join(facesDir, fmt.Sprintf("face_%d.jpg", i+1))
				saveFaceImage(originalImage, face.Rectangle, faceImageFile)
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
		fileInfos, err := os.ReadDir(facesDir)
		if err != nil {
			http.Error(w, "Error al leer las imágenes de caras encontradas", http.StatusInternalServerError)
			return
		}

		var carasDetectadas []string
		for _, fileInfo := range fileInfos {
			if strings.HasSuffix(fileInfo.Name(), ".jpg") {
				carasDetectadas = append(carasDetectadas, fileInfo.Name())
			}
		}

		tpl, err := template.New("result").Parse(`
			<h1>Caras detectadas</h1>
			{{range $index, $label := .}}
			<img src="/faces/{{.}}" alt="Persona {{$index}}">
			<label>Persona {{$index}}</label>
			<br>
			{{end}}
			<br>
			<a href="/">Volver</a>
		`)
		if err != nil {
			http.Error(w, "Error al renderizar el resultado", http.StatusInternalServerError)
			return
		}

		tpl.Execute(w, carasDetectadas)
	})

	// Servir las caras encontradas.
	http.Handle("/faces/", http.StripPrefix("/faces/", http.FileServer(http.Dir(facesDir))))

	// Iniciar el servidor en el puerto 8080.
	log.Println("El servidor está ejecutándose en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Función para guardar la imagen de la cara detectada.
func saveFaceImage(originalImage image.Image, rect image.Rectangle, filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Convertir la imagen decodificada a *image.RGBA.
	rgbaImage := image.NewRGBA(rect)
	draw.Draw(rgbaImage, rgbaImage.Bounds(), originalImage, rect.Min, draw.Src)

	// Recortar la región de la cara de la imagen original.
	croppedImage := rgbaImage.SubImage(rect).(*image.RGBA)

	// Guardar la imagen recortada en formato JPEG.
	return jpeg.Encode(outFile, croppedImage, nil)
}
