package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type newSong struct{
	Login string `json:"login"`
	Password string `json:"password"`
	Name string `json:"name"`
	Mp3File []byte `json:"file"`
}

type newName struct{
	Login string `json:"login"`
	Password string `json:"password"`
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type User struct{
	Login string `json:"login"`
	Password string `json:"password"`
}

type Answer struct{
	Login string `json:"login"`
	Names []string   `json:"names"`
	Mp3Files [][]byte `json:"files"`
}

func main() {
    //getInApp("http://127.0.0.1:8080/getInApp/registration","POST","m6", "1243")
 //   changeName("http://127.0.0.1:8080/audio/changeName","m6","1243", "music0010","newName")
	addAudio("http://127.0.0.1:8080/audio/addSong", "D:\\music/mus3.mp3","m6","1243","music0010")
//	getAudios("http://127.0.0.1:8080/audio/getNamesOfSongs", "m6","1243")
}

func getInApp(url string, method string, login string, password string){
	data, _ := json.Marshal(User{
		Login:    login,
		Password: password,
	})
	reader := bytes.NewReader(data)
	req, err := http.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func getAudios(path string, login string, password string){
	data, _ := json.Marshal(User{
		Login:    login,
		Password: password,
	})
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", path, reader)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in client")
		return
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	var answer Answer
	err = json.Unmarshal(body, &answer)
	if err != nil {
		log.Println("Error in unmarshal")
	} else {
		os.RemoveAll("D:\\users/"+answer.Login)
		err = os.MkdirAll("D:\\users/"+answer.Login, 0777)
		if err != nil {
			log.Println(err)
			return
		}
		for index, elem := range answer.Mp3Files {
			err:=ioutil.WriteFile("D:\\users/"+answer.Login+"/"+answer.Names[index]+".mp3",elem,0644)
			if err!=nil{
				log.Println(err)
				return
			}
		}
	}
}

func changeName(path string, login string, password string, oldNameOfSong string, newNameOfSong string){
	data, _ := json.Marshal(newName{
		Login:login,
		Password:password,
		OldName:oldNameOfSong,
		NewName:newNameOfSong,
	})
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", path, reader)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in client")
		return
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func addAudio(path string, fileName string, login string, password string, nameOfSong string) {
	file, err := ioutil.ReadFile(fileName)
	if err!=nil{
		log.Println("Error in reading file")
		return
	}
	data, _ := json.Marshal(newSong{
		Login:    login,
		Password: password,
		Name:     nameOfSong,
		Mp3File:  file,
	})
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", path, reader)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in client")
		return
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
