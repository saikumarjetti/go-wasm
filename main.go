package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"syscall/js"

	"github.com/auyer/steganography"
)

func valueToByteArray(v js.Value) []byte {
	binImage := make([]byte, v.Length())
	js.CopyBytesToGo(binImage, v)

	return binImage
}
func Encrypt(plaintext []byte, key []byte) (ciphertext []byte, err error) {
	k := sha256.Sum256(key)
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key []byte) (plaintext []byte, err error) {
	k := sha256.Sum256(key)
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// OpenImageFile open a image and return a image object
func OpenImageFile(imgByte []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		return nil, err
	}

	return img, nil
}
func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// func add(this js.Value, i []js.Value) interface{} {
// 	println(i)
// 	return js.ValueOf(i[0].Int() + i[1].Int())
// }
func imageEncode(this js.Value, i []js.Value) interface{} {

	imgArr := i[0]
	inBuf := make([]uint8, imgArr.Get("byteLength").Int())
	js.CopyBytesToGo(inBuf, imgArr)
	img, _ := OpenImageFile(inBuf)

	w := new(bytes.Buffer) // buffer that will recieve the results
	mssg := []byte(i[1].String())
	//
	key := []byte(i[2].String())
	ct, err := Encrypt([]byte(mssg), key)
	if err != nil {
		panic(err)
	}
	EncryptText := base64.StdEncoding.EncodeToString(ct)
	fmt.Println("Encrypted:", base64.StdEncoding.EncodeToString(ct))
	//
	errImg := steganography.Encode(w, img, []byte(EncryptText))
	if errImg != nil {
		log.Printf("Error Encoding file %v", err)
		return ""
	}

	oi, _ := png.Decode(w)
	var buff bytes.Buffer
	png.Encode(&buff, oi)
	encodeString := base64.StdEncoding.EncodeToString(buff.Bytes())

	return js.ValueOf(encodeString)
}
func imageDecode(this js.Value, i []js.Value) interface{} {
	imgArr := i[0]
	inBuf := make([]uint8, imgArr.Get("byteLength").Int())
	js.CopyBytesToGo(inBuf, imgArr)
	img, _ := OpenImageFile(inBuf)
	// img, _ = png.Decode(inBuf)
	pass := i[1].String()

	sizeOfMessage := steganography.GetMessageSizeFromImage(img) // retrieving message size to decode in the next line
	fmt.Println(sizeOfMessage)
	msg := steganography.Decode(sizeOfMessage, img) // decoding the message from the file
	data1, err1 := base64.StdEncoding.DecodeString(string(msg))
	if err1 != nil {
		fmt.Println("error:", err1)
	}
	pt, err := Decrypt(data1, []byte(pass))
	if err != nil {
		panic(err)
	}
	return js.ValueOf(string(pt))
}

func registerCallbacks() {
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("imageEncode", js.FuncOf(imageEncode))
	js.Global().Set("imageDecode", js.FuncOf(imageDecode))

}

// exposing to JS

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}

// GOARCH=wasm GOOS=js go build -o lib.wasm main.go
