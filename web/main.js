document.addEventListener("DOMContentLoaded", () => {
    const adofaiInput = document.getElementById("adofaiFile")
    const uploadProgress = document.getElementById("uploadProgress")
    let uploadIndex = 0

    function upload(input, reader, index) {
        uploadIndex = index
        if (index >= input.files.length) {
            input.value = ""
            return
        }

        reader.readAsArrayBuffer(input.files[index])
    }

    function uploadByChunk(sessionID, chunkSize, byteArray) {
        console.log(byteArray.byteLength)
        const chunkCount = Math.ceil(byteArray.byteLength / chunkSize)
        if (chunkCount === 0) return

        let chunkUploadComplete = false

        let chunkIndex = 0
        function uploadNext() {
            let data = byteArray.slice(chunkIndex * chunkSize, (chunkIndex + 1) * chunkSize)

            chunkUploadComplete = true
            fetch(`/upload/${sessionID}`, {
                method: "POST",
                body: data
            })
                .then((response) => response.text())
                .then((result) => {
                    if (result === "Success") {
                        chunkUploadComplete = true
                        uploadProgress.value = ++chunkIndex / chunkCount * 100

                        if (chunkIndex < chunkCount) {
                            uploadNext()
                        } else {
                            fetch(`/upload/${sessionID}`, {
                                method: "DELETE"
                            })
                                .then((response) => response.text())
                                .then((result) => {
                                    uploadProgress.value = 0
                                    console.log(result)
                                })
                                .then(() => {
                                    fetch(`/get_download/${sessionID}`, {
                                        method: "GET"
                                    })
                                        .then((response) => response.text())
                                        .then((result) => {
                                            let url = `https://download.aef.kr/${result}`;
                                            document.write(`Upload Success!<br>Download: <a href="${url}">${url}</a>`)
                                        })
                                })
                        }
                    }
                })
        }
        uploadNext()
    }

    adofaiInput.addEventListener("change", (event) => {
        const input = event.target
        const reader = new FileReader()

        reader.onload = () => {
            fetch("/get_session", {
                method: "GET"
            })
                .then((response) => response.json())
                .then((result) => {
                    uploadByChunk(result["sessionID"], result["chunkSize"], reader.result)
                })
            upload(input, reader, uploadIndex + 1)
        }

        upload(input, reader, 0)
    })
})