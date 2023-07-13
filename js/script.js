const logregBox= document.querySelector('.logreg-box');
const loginLink= document.querySelector('.login-link');
const registerLink= document.querySelector('.register-link');

registerLink.addEventListener('click',() =>{
logregBox.classList.add('active');

});


loginLink.addEventListener('click',() =>{
    logregBox.classList.remove('active');
    
    });


    // Función para enviar una imagen al servidor
function enviarImagen(imageFile) {
    const url = "http://localhost:8080/detect"; // Actualiza con la URL correcta del backend
    
    const formData = new FormData();
    formData.append("image", imageFile);
  
    return fetch(url, {
      method: "POST",
      body: formData
    })
    .then(response => response.json())
    .then(data => {
      // Aquí puedes procesar los resultados recibidos del backend
      console.log("Resultados:", data);
    })
    .catch(error => {
      console.error("Error al enviar la imagen:", error);
    });
  }
  
  // Manejar el evento de cambio del input de archivos
  const inputFile = document.querySelector("input[type='file']");
  inputFile.addEventListener("change", (event) => {
    const file = event.target.files[0];
    if (file) {
      enviarImagen(file);
    }
  });
  