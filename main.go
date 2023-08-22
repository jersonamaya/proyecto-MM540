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
	"time"

	"image"
	"image/draw"
	"image/jpeg"

	"github.com/Kagami/go-face"
	"github.com/nfnt/resize"
)

const dataDir = "testdata"
const facesDetectedDir = "faces_detected"
const templatesDir = "templates"

var modelsDir = filepath.Join(dataDir, "models")
var imagesDir = filepath.Join(dataDir, "images")

var faceToOriginalImageMap = make(map[string]string)

type TemplateData struct {
	Label      string
	FaceImages []string
}

func main() {
	rec, err := face.NewRecognizer(modelsDir)
	if err != nil {
		log.Fatalf("No se puede inicializar el reconocedor facial: %v", err)
	}
	defer rec.Close()

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"listFacesInFolder": listFacesInFolder,
	}).ParseGlob(filepath.Join(templatesDir, "/*.html")))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			file, _, err := r.FormFile("image")
			if err != nil {
				http.Error(w, "Error al cargar la imagen", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			timestamp := fmt.Sprintf("%d", nowUnixNano())
			imageFolder := filepath.Join(facesDetectedDir, "image_"+timestamp)
			err = os.MkdirAll(imageFolder, 0755)
			if err != nil {
				http.Error(w, "Error al crear la carpeta de la imagen", http.StatusInternalServerError)
				return
			}

			originalImageFile := filepath.Join(imageFolder, "original.jpg")
			originalFile, err := os.Create(originalImageFile)
			if err != nil {
				http.Error(w, "Error al guardar la imagen original", http.StatusInternalServerError)
				return
			}
			defer originalFile.Close()
			_, err = io.Copy(originalFile, file)
			if err != nil {
				http.Error(w, "Error al guardar la imagen original", http.StatusInternalServerError)
				return
			}

			faces, err := rec.RecognizeFile(originalImageFile)
			if err != nil {
				http.Error(w, "Error al reconocer las caras", http.StatusInternalServerError)
				return
			}

			for i, face := range faces {
				faceImageFile := filepath.Join(imageFolder, fmt.Sprintf("face_%d.jpg", i+1))
				saveFaceImage(originalImageFile, face.Rectangle, faceImageFile)
				faceToOriginalImageMap[filepath.Base(faceImageFile)] = filepath.Base(originalImageFile)
			}

			http.Redirect(w, r, "/result", http.StatusSeeOther)
		} else {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "subir.html", nil)
	})

	http.HandleFunc("/subir", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "subir.html", nil)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		fileInfos, err := os.ReadDir(facesDetectedDir)
		if err != nil {
			http.Error(w, "Error al leer las imágenes de caras detectadas", http.StatusInternalServerError)
			return
		}

		var carasDetectadas []string
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				carasDetectadas = append(carasDetectadas, fileInfo.Name())
			}
		}

		data := []TemplateData{}
		for _, label := range carasDetectadas {
			faceImages, _ := listFacesInFolder(label)
			data = append(data, TemplateData{Label: label, FaceImages: faceImages})
		}

		tmpl.ExecuteTemplate(w, "result.html", data)
	})

	http.Handle("/faces_detected/", http.StripPrefix("/faces_detected/", http.FileServer(http.Dir(facesDetectedDir))))

	log.Println("El servidor está ejecutándose en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func saveFaceImage(originalImagePath string, rect image.Rectangle, filename string) error {
	img, err := os.Open(originalImagePath)
	if err != nil {
		return err
	}
	defer img.Close()

	originalImage, _, err := image.Decode(img)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	rgbaImage := image.NewRGBA(rect)
	draw.Draw(rgbaImage, rgbaImage.Bounds(), originalImage, rect.Min, draw.Src)

	const fixedWidth, fixedHeight = 200, 200
	resizedImage := resize.Resize(fixedWidth, fixedHeight, rgbaImage, resize.Lanczos3)

	return jpeg.Encode(outFile, resizedImage, nil)
}

func listFacesInFolder(folderName string) ([]string, error) {
	folderPath := filepath.Join(facesDetectedDir, folderName)
	fileInfos, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var faceImages []string
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		if strings.HasPrefix(fileInfo.Name(), "face_") && strings.HasSuffix(fileInfo.Name(), ".jpg") {
			faceImages = append(faceImages, fileInfo.Name())
		}
	}

	return faceImages, nil
}

func nowUnixNano() int64 {
	return time.Now().UnixNano()
}
