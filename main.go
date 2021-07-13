package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	getInApp("http://127.0.0.1:8080/getInApp/enter","POST",[]byte(`{
    "loginOfUser": "m4",
    "passwordOfUser": "1243"
}`))
	sendAudio("http://127.0.0.1:8080/audio", "D:\\mus.mp3","m4","1243","music")
}

func getInApp(url string, method string, data [] byte){
	reader := bytes.NewReader(data)
	req, err := http.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func sendAudio(ip string, fileName string, login string, password string, nameOfSong string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	var requestBody bytes.Buffer
	multiPartWriter := multipart.NewWriter(&requestBody)
	fileWriter, err := multiPartWriter.CreateFormFile("audioFile", fileName)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalln(err)
	}
	fieldWriter, err := multiPartWriter.CreateFormField("login")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = fieldWriter.Write([]byte(login))
	if err != nil {
		log.Fatalln(err)
	}
	fieldWriter, err = multiPartWriter.CreateFormField("password")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = fieldWriter.Write([]byte(password))
	if err != nil {
		log.Fatalln(err)
	}
	fieldWriter, err = multiPartWriter.CreateFormField("nameOfSong")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = fieldWriter.Write([]byte(nameOfSong))
	if err != nil {
		log.Fatalln(err)
	}
	multiPartWriter.Close()
	req, err := http.NewRequest("POST", ip, &requestBody)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}