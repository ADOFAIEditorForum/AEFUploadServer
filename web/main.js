document.addEventListener("DOMContentLoaded", () => {
    const adofaiInput = document.getElementById("adofaiFile")
    let uploadIndex = 0

    function upload(input, reader, index) {
        uploadIndex = index
        if (index >= input.files.length) {
            input.value = ""
            return
        }

        reader.readAsArrayBuffer(input.files[index])
    }

    adofaiInput.addEventListener("change", (event) => {
        const input = event.target
        const reader = new FileReader()

        reader.onload = () => {
            fetch("/", {
                method: "POST",
                body: reader.result
            })
                .then((response) => response.text())
                .then((result) => console.log(result))
            upload(input, reader, uploadIndex + 1)
        }

        upload(input, reader, 0)
    })
})