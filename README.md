# Go-WASM Secure Image Steganography

## Description

This project implements a secure method for hiding data within images (steganography) directly in the web browser using Go compiled to WebAssembly (WASM). It provides two-level security by first encrypting the secret message with AES and then compressing it with LZW before embedding it into the cover image using the Least Significant Bit (LSB) algorithm.

The primary goal is to enhance data security and privacy by performing all computationally intensive and sensitive operations (encryption, compression, steganography) on the client-side, eliminating the need to send user data or images to a server for processing. This leverages the near-native performance of WebAssembly for tasks traditionally difficult to perform efficiently or securely in standard web applications.

## Features

* **Client-Side Processing:** All core operations run in the user's browser using WebAssembly.
* **Two-Level Security:**
    * **AES Encryption:** Encrypts the secret message using a user-provided password.
    * **LZW Compression:** Compresses the encrypted data to maximize storage capacity within the image (lossless compression).
* **LSB Steganography:** Hides the compressed, encrypted data within the least significant bits of the cover image's pixels.
* **Go & WebAssembly:** Core logic written in Go and compiled to WASM for performance.
* **Web Interface:** Includes HTML, CSS, and JS for user interaction (image selection, message input, password, encode/decode actions).
* **Go Web Server:** A simple backend server (`server.go`) primarily to serve the static files (HTML, JS, CSS, WASM).

## Getting Started

### Prerequisites

* **Go:** Version 1.11+ installed ([https://golang.org/dl/](https://golang.org/dl/)).
* **Web Browser:** Modern browser with WebAssembly support.

### Building the WebAssembly Module

1.  Clone the repository:
    ```bash
    git clone [https://github.com/saikumarjetti/go-wasm.git](https://github.com/saikumarjetti/go-wasm.git)
    cd go-wasm
    ```
2.  Compile the Go code (`main.go`) to WASM (`lib.wasm`):
    ```bash
    GOOS=js GOARCH=wasm go build -o lib.wasm main.go
    ```

### Running the Application

1.  Start the local Go web server:
    ```bash
    go run server.go
    ```
2.  Open your browser and navigate to `http://localhost:8080` (or the port specified by the server output).

## How it Works

### Encoding

1.  User provides a cover image, a secret message, and a password via the web interface.
2.  The message is encrypted using AES with the provided password.
3.  The encrypted data is compressed using the LZW algorithm.
4.  The compressed data is embedded into the pixels of the cover image using the LSB technique.
5.  A new image (stego-image) containing the hidden data is generated and offered for download.

### Decoding

1.  User provides the stego-image and the correct password used during encoding.
2.  The hidden data is extracted from the image's least significant bits.
3.  The extracted data is decompressed using the LZW algorithm.
4.  The resulting encrypted data is decrypted using AES with the provided password.
5.  The original secret message is displayed to the user.

## Key Files

* `main.go`: Go source containing the AES, LZW, and LSB logic compiled to WASM.
* `lib.wasm`: The compiled WebAssembly module.
* `wasm_exec.js`: Go-provided JavaScript file to load and run the WASM module.
* `server.go`: Simple Go HTTP server to serve the application files.
* `index.html` / `decode.html`: HTML pages for the user interface.
* `script.js`: JavaScript for interacting with the UI and the WASM module.
* `style.css`: Styles for the web interface.
* `go.mod` / `go.sum`: Go module dependency files.
