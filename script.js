const list= document.querySelectorAll('.list');

function activeLink(){
    list.forEach((item)=>
    item.classList.remove('active'));
    this.classList.add('active');
}

list.forEach((item)=>
item.addEventListener('click',activeLink));
document.addEventListener('DOMContentLoaded', () => {
    const uploadForm = document.getElementById('upload-form');

    if (uploadForm) {
        uploadForm.addEventListener('submit', async (event) => {
            event.preventDefault();

            const formData = new FormData(uploadForm);

            try {
                const response = await fetch('http://localhost:8080/upload', {
                    method: 'POST',
                    body: formData,
                });

                if (response.ok) {
                    const result = await response.text();
                    console.log(result); // Imprime "Imagen cargada y rostros detectados"
                } else {
                    console.error('Error en la solicitud:', response.status);
                }
            } catch (error) {
                console.error('Hubo un error:', error);
            }
        });
    }
});


// Obtén una referencia al enlace "Ver Resultados"
const resultLink = document.querySelector('a[href="/result"]');

// Agrega un evento de clic al enlace
resultLink.addEventListener('click', (event) => {
  event.preventDefault(); // Evita que el enlace realice la redirección por defecto
  window.location.href = 'http://localhost:8080/result'; // Redirige a la página de resultados
});


document.getElementById('image').addEventListener('change', function(event) {
    const previewImage = document.getElementById('preview-image');
    previewImage.src = URL.createObjectURL(event.target.files[0]);
});





