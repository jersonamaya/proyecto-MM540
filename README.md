# proyecto-MM540
El proyecto consiste en el desarrollo de un sistema web de almacenamiento de fotografías que incluye diversas herramientas de procesamiento de imágenes. Una de las funcionalidades destacadas es la auto-clasificación de las fotografías de acuerdo al rostro de las personas presentes en ellas.

El sistema permitirá a los usuarios cargar sus fotografías en la plataforma, ya sea a través de una interfaz web o mediante la sincronización con aplicaciones móviles. Una vez cargadas, el sistema utilizará algoritmos de reconocimiento facial para identificar y reconocer los rostros en cada fotografía.

La funcionalidad de auto-clasificación se basa en estos algoritmos de reconocimiento facial. El sistema analizará los rostros identificados en las fotografías y los asociará con personas previamente registradas en la plataforma. Por ejemplo, si una persona llamada "Jose" aparece en varias fotografías, el sistema agrupará automáticamente todas las fotos en una categoría o álbum etiquetado como "Jose". Esto simplifica la organización y búsqueda de las fotografías para el usuario, ya que puede acceder rápidamente a todas las imágenes de una persona específica.

 este proyecto utiliza la liberia  [github.com/Kagami/go-face](https://github.com/Kagami/go-face) que es  go-face implementa el reconocimiento facial para Go usando [dlib](http://dlib.net/) , un popular conjunto de herramientas de aprendizaje automático.[ Lea el artículo Reconocimiento facial con Go ](https://hackernoon.com/face-recognition-with-go-676a555b8a7e) para obtener algunos detalles básicos si es nuevo en el concepto [FaceNet ](https://arxiv.org/abs/1503.03832).

 
 # Requisitos
 
Para compilar go-face necesita tener instalados los paquetes de desarrollo dlib (>= 19.10) y libjpeg.

# Ubuntu 18.10+, lado de Debian
Las últimas versiones de Ubuntu y Debian proporcionan el paquete dlib adecuado, así que simplemente ejecute:

```cmd
# Ubuntu
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg-turbo8-dev
# Debian
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg62-turbo-dev
```
## Mac OS
Asegúrate de tener [ Homebrew](https://brew.sh/) instalado.
```cmd
brew install dlib
```
#Windows 
--------              
Asegúrese de tener [MSYS2](https://www.msys2.org/) instalado.

1.Ejecutar ```cmd MSYS2``` MSYSshell desde el menú Inicio
2. Ejecute ```cmd  pacman -Syuy``` si le pide que cierre el shell, hágalo.
3.Corre ```cmd pacman -Syu``` de nuevo
4.Correr ```cmd pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-dlib```
5.  i.Si ya tiene Go y Git instalados y disponibles en ```cmd set MSYS2_PATH_TYPE=inherit```la línea de comentario de PATH ```cmd msys2_shell.cmd``` ubicada en la carpeta de instalación 
    de MSYS2
 ii.De lo contrario, ejecute ```cmd pacman -S mingw-w64-x86_64-go git```
6. Ejecute ```cmd MSYS2 MinGW 64-bitshell``` desde el menú Inicio para compilar y usar go-face
                
-------
  # Otros sistemas
Intente instalar dlib/libjpeg con el administrador de paquetes de su distribución o [compílelo desde las fuentes](http://dlib.net/compile.html) . Tenga en cuenta que go-face no funcionará con paquetes antiguos de dlib como libdlib18. Alternativamente, cree un problema con el nombre de su sistema y alguien podría ayudarlo con el proceso de instalación.               
# Modelos
Actualmente ```cmd shape_predictor_5_face_landmarks.dat ```y ```cmd mmod_human_face_detector.dat```son ```cmd dlib_face_recognition_resnet_model_v1.dat``` obligatorios. Puede descargarlos desde el repositorio [go-face-testdata](https://github.com/Kagami/go-face-testdata) :
```cmd
wget https://github.com/Kagami/go-face-testdata/raw/master/models/shape_predictor_5_face_landmarks.dat
wget https://github.com/Kagami/go-face-testdata/raw/master/models/dlib_face_recognition_resnet_model_v1.dat
wget https://github.com/Kagami/go-face-testdata/raw/master/models/mmod_human_face_detector.dat
```
# uso
Para usar go-face en su código Go:

import "github.com/Kagami/go-face" 
Para instalar go-face en tu $GOPATH:

```cmd go get github.com/Kagami/go-face```
Para obtener más detalles, consulte [la documentación de GoDoc](https://pkg.go.dev/github.com/Kagami/go-face).

  # Ejemplo
``` golang
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Kagami/go-face"
)

// Path to directory with models and test images. Here it's assumed it
// points to the <https://github.com/Kagami/go-face-testdata> clone.
const dataDir = "testdata"

var (
	modelsDir = filepath.Join(dataDir, "models")
	imagesDir = filepath.Join(dataDir, "images")
)

// This example shows the basic usage of the package: create an
// recognizer, recognize faces, classify them using few known ones.
func main() {
	// Init the recognizer.
	rec, err := face.NewRecognizer(modelsDir)
	if err != nil {
		log.Fatalf("Can't init face recognizer: %v", err)
	}
	// Free the resources when you're finished.
	defer rec.Close()

	// Test image with 10 faces.
	testImagePristin := filepath.Join(imagesDir, "pristin.jpg")
	// Recognize faces on that image.
	faces, err := rec.RecognizeFile(testImagePristin)
	if err != nil {
		log.Fatalf("Can't recognize: %v", err)
	}
	if len(faces) != 10 {
		log.Fatalf("Wrong number of faces")
	}

	// Fill known samples. In the real world you would use a lot of images
	// for each person to get better classification results but in our
	// example we just get them from one big image.
	var samples []face.Descriptor
	var cats []int32
	for i, f := range faces {
		samples = append(samples, f.Descriptor)
		// Each face is unique on that image so goes to its own category.
		cats = append(cats, int32(i))
	}
	// Name the categories, i.e. people on the image.
	labels := []string{
		"Sungyeon", "Yehana", "Roa", "Eunwoo", "Xiyeon",
		"Kyulkyung", "Nayoung", "Rena", "Kyla", "Yuha",
	}
	// Pass samples to the recognizer.
	rec.SetSamples(samples, cats)

	// Now let's try to classify some not yet known image.
	testImageNayoung := filepath.Join(imagesDir, "nayoung.jpg")
	nayoungFace, err := rec.RecognizeSingleFile(testImageNayoung)
	if err != nil {
		log.Fatalf("Can't recognize: %v", err)
	}
	if nayoungFace == nil {
		log.Fatalf("Not a single face on the image")
	}
	catID := rec.Classify(nayoungFace.Descriptor)
	if catID < 0 {
		log.Fatalf("Can't classify")
	}
	// Finally print the classified label. It should be "Nayoung".
	fmt.Println(labels[catID])
} 
```
Corre con:
mkdir -p ~/go && cd ~/go  # Or cd to your $GOPATH
mkdir -p src/go-face-example && cd src/go-face-example
git clone https://github.com/Kagami/go-face-testdata testdata
edit main.go  # Paste example code
``` go get && go run main.g```

 # Prueba
Para recuperar datos de prueba y ejecutar pruebas:

``` make test```          
