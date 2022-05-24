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
	"strconv"
	"strings"
	"syscall/js"

	"github.com/auyer/steganography"
)

var debuging bool = true

func compressLZW(testStr string) []int {
	code := 256
	dictionary := make(map[string]int)
	for i := 0; i < 256; i++ {
		dictionary[string(rune(i))] = i
	}
	currChar := ""
	result := make([]int, 0)
	for _, c := range []byte(testStr) {
		phrase := currChar + string(c)
		if _, isTrue := dictionary[phrase]; isTrue {
			currChar = phrase
		} else {
			result = append(result, dictionary[currChar])
			dictionary[phrase] = code
			code++
			currChar = string(c)
		}
	}
	if currChar != "" {
		result = append(result, dictionary[currChar])
	}
	return result
}

func decompressLZW(compressed []int) string {
	code := 256
	dictionary := make(map[int]string)
	for i := 0; i < 256; i++ {
		dictionary[i] = string(i)
	}

	currChar := string(compressed[0])
	result := currChar
	for _, element := range compressed[1:] {
		var word string
		if x, ok := dictionary[element]; ok {
			word = x
		} else if element == code {
			word = currChar + currChar[:1]
		} else {
			panic(fmt.Sprintf("Bad compressed element: %d", element))
		}

		result += word

		dictionary[code] = currChar + word[:1]
		code++

		currChar = word
	}
	return result
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

func imgArrToGoImage(imgArr js.Value) (image.Image, error) {

	imgBuff := make([]uint8, imgArr.Get("byteLength").Int()) // creating a uint buffer to store js image array
	js.CopyBytesToGo(imgBuff, imgArr)                        // copies bytes from src(imgArr) to dst(imgBuff)
	img, _, err := image.Decode(bytes.NewReader(imgBuff))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func imageEncode(this js.Value, i []js.Value) interface{} {

	imgArr := i[0] // image arr from frontend
	img, _ := imgArrToGoImage(imgArr)
	w := new(bytes.Buffer)  // buffer that will recieve the results from stegno
	mssg := (i[1].String()) // user message from front end
	if debuging {
		fmt.Println("mssg = ")
		fmt.Println(mssg)
	}

	// compressing the data using LZW o/p is int[]
	compressedData := compressLZW(mssg)
	if debuging == true {
		fmt.Println("compressed data")
		fmt.Println(compressedData)
	}

	// cnverting int[] to string to encrypt
	stringData := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(compressedData)), ","), "[]")

	key := []byte(i[2].String())                //coverting password to byte[]
	ct, err := Encrypt([]byte(stringData), key) // encrypting the message with the providede password
	if err != nil {
		panic(err)
	}
	if debuging {
		fmt.Println("EncryptText :")
		fmt.Println(base64.StdEncoding.EncodeToString(ct))
	}
	errImg := steganography.Encode(w, img, []byte(ct))
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

	imgArr := i[0] // image arr from frontend
	img, _ := imgArrToGoImage(imgArr)
	// img, _ = png.Decode(imgBuff)

	password := i[1].String() // password from frontend and converting it into string so go can understand

	// gets the size of the message from the first four bytes encoded in the image
	sizeOfMessage := steganography.GetMessageSizeFromImage(img)

	if debuging {

		fmt.Print("sizeOfMessage = ")
		fmt.Println(sizeOfMessage)
	}

	msg := steganography.Decode(sizeOfMessage, img) // decoding the message from the file

	if debuging {

		fmt.Println("extracted encrypted msg from image")
		fmt.Println(base64.StdEncoding.EncodeToString(msg))
	}

	pt, err := Decrypt(msg, []byte(password))
	if err != nil {
		panic(err)
	}

	if debuging {
		fmt.Println("DecryptedText : ")
		fmt.Println(base64.StdEncoding.DecodeString(string(msg)))
	}

	// after decryption we get compressed data in one stringseparated by ','
	// converting string to string arr[] on ","
	strArryData := strings.Split(string(pt), ",")
	var LZWIntData = []int{}

	if debuging {

		fmt.Println("strArryData")
		fmt.Println(strArryData)
	}

	// string num array to int[]
	for _, i := range strArryData {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		LZWIntData = append(LZWIntData, j)
	}
	decompressedData := decompressLZW(LZWIntData)
	fmt.Println("decompressedData")
	fmt.Println(decompressedData)
	fmt.Println(password)
	return js.ValueOf(string(decompressedData))
}

// exposing to JS
func registerCallbacks() {
	js.Global().Set("imageEncode", js.FuncOf(imageEncode))
	js.Global().Set("imageDecode", js.FuncOf(imageDecode))
}

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	// to stop go from closing
	<-c
}

// GOARCH=wasm GOOS=js go build -o lib.wasm main.go
