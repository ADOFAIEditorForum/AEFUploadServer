document.addEventListener("DOMContentLoaded", () => {
    const adofaiInput = document.getElementById("adofaiFile")

    adofaiInput.addEventListener("change", (event) => {
        const input = event.target
        const reader = new FileReader()

        reader.onload = () => {
            fetch("/upload", {
                method: "POST",
                body: reader.result
            })
                .then((response) => response.text())
                .then((result) => console.log(result))
        }

        reader.readAsArrayBuffer(input.files[0])
    })
})