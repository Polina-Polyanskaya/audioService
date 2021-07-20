package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const dbName string = "DataBase"
const colName string = "User"

type User struct {
	Login string `json:"login"`
	Password string `json:"password"`
	Names [] string `json:"namesOfSongs"`
}

type newName struct{
	Login string `json:"login"`
	Password string `json:"password"`
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type Answer struct{
	Login string `json:"login"`
	Names []string   `json:"names"`
	Mp3Files [][]byte `json:"files"`
}

type Audio struct{
	Login string `json:"login"`
	Password string `json:"password"`
	Name string `json:"name"`
	Mp3File []byte `json:"file"`
}

func main() {
	handler := http.NewServeMux()
	handler.HandleFunc("/getInApp/", Logger(authHandler))
	handler.HandleFunc("/audio/", Logger(audioHandler))
	s := http.Server{
		Addr: ":8080",
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}

//аутенфикация
func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//получаем путь
	name := strings.Replace(r.URL.Path, "/getInApp/", "", 1)
	//проверка метода
	if r.Method != http.MethodPost {
		//метод неправильный
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//проверка, что путь правильный
	if !(name == "registration" || name == "enter") {
		//путь неправильный
		http.Error(w, "Wrong path", 418)
		return
	}
	var user User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	user.Names = []string{}
	//проверка, правильный ли json был передан и можно ли создать User
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//проверка значений объекта User
	if !checkFieldsToAuth(user, w) {
		//проверка соединения с бд mongoDb
		wasError, client := wasErrorInConnectToDb()
		if wasError {
			//ошибка подключения к бд
			w.WriteHeader(http.StatusNotFound)
			return
		}
		isInTable, _ := isUserInTable(client, user)
		//если путь registration
		if name == "registration" {
			//проверка на наличие такого пользователя в бд
			if isInTable {
				//если такой пользователь есть уже
				http.Error(w, "Already exist", 418)
				return
			}
			//если такого пользователя нет, то добавить в бд
			if addToTable(client, user) {
				http.Error(w, "Was not added to db", 418)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
		//если путь enter
		if name == "enter" {
			//если такого пользователя уже есть в бд
			if isInTable {
				w.WriteHeader(http.StatusOK)
			} else {
				//если такого пользователя еще нет в бд
				http.Error(w, "Not exist", 418)
			}
		}
	}
}

//проверка значений логина и пароля
func checkFieldsToAuth(user User, w http.ResponseWriter) bool {
	//проверка логина
	if isErrorInLogin(user.Login, w) {
		return true
	} else if isErrorInPassword(user.Password, w) { //проверка пароля
		return true
	}
	return false
}

func isErrorInLogin(login string, w http.ResponseWriter) bool {
	if login == "" {
		http.Error(w, "Wrong login", 418)
		return true
	}
	return false
}

func isErrorInPassword(password string, w http.ResponseWriter) bool {
	if password == "" {
		http.Error(w, "Wrong password", 418)
		return true
	}
	return false
}

//соединение к mongodb
func wasErrorInConnectToDb() (bool, *mongo.Client) {
	wasError := false
	//clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://Polina:Polina1521Misha@cluster0.v80ex.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		wasError = true
	} else {
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			wasError = true
		}
	}
	if !wasError {
		log.Printf("Connected to MongoDB!")
	}else{
		log.Printf("Did not connect to MongoDB!")
	}
	return wasError, client
}

//добавление нового пользоваателя в таблицу
func addToTable(client *mongo.Client, user User) bool{
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := client.Database(dbName).Collection(colName)
	result, insertErr := col.InsertOne(ctx, user)
	if insertErr != nil {
		log.Printf("InsertOne ERROR:%v", insertErr)
		return true
	} else {
		newID := result.InsertedID
		log.Printf("InsertedOne(), newID", newID)
		return false
	}
}

func uploadTable(w http.ResponseWriter, client *mongo.Client, audio Audio) {
	checker, names:=isUserInTable(client, User{
		Login: audio.Login,
		Password: audio.Password,
		Names: nil,
	})
	if !checker {
		http.Error(w, "Not exist", 418)
		return
	}
	checker = false
	for _, elem := range names {
		if elem == audio.Name {
			checker = true
		}
	}
	if checker {
		http.Error(w, "Already exist", 418)
		return
	}
	col := client.Database(dbName).Collection(colName)
	filter := bson.D{{"login", audio.Login}, {"password", audio.Password}}
	update := bson.D{
		{"$push", bson.D{
			{"names", audio.Name},
		}},
	}
	_, err := col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println(err)
		http.Error(w, "Did not upload", 418)
	}
	err = os.MkdirAll("D:\\Data/"+audio.Login, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Did not upload", 418)
		return
	}
	err=ioutil.WriteFile("D:\\Data/"+audio.Login+"/"+audio.Name+".mp3",audio.Mp3File,0644)
	if err!=nil{
		log.Println(err)
		http.Error(w, "Did not upload", 418)
		return
	}
	w.WriteHeader(200)
}

func isErrorInName(name string, w http.ResponseWriter) bool{
	if name == "" {
		http.Error(w, "Wrong name", 418)
		return true
	}
	return false
}

func checkFormValues(login string, password string, name string, w http.ResponseWriter) bool{
	if isErrorInLogin(login, w){
		return true
	} else if isErrorInPassword(password, w) {//проверка пароля
		return true
	} else if isErrorInName(name, w) { //проверка имени песни
		return true
	}
	return false
}

func changeName(w http.ResponseWriter, client *mongo.Client, nameChanger newName) {
	checker, names := isUserInTable(client, User{
		Login:    nameChanger.Login,
		Password: nameChanger.Password,
		Names:    nil,
	})
	if !checker {
		http.Error(w, "Not exist", 418)
		return
	}
	checker = false
	alreadyHasNewName:=false
	for _, elem := range names {
		if elem == nameChanger.OldName {
			checker = true
		}
		if elem == nameChanger.NewName {
			alreadyHasNewName=true
		}
	}
	if !checker {
		http.Error(w, "Not exist oldName", 418)
		return
	}
	if alreadyHasNewName{
		http.Error(w, "Already exist newName", 418)
		return
	}
	col := client.Database(dbName).Collection(colName)
	filter := bson.D{{"login", nameChanger.Login}, {"password", nameChanger.Password},{"names", nameChanger.OldName}}
	update := bson.D{
		{"$set", bson.D{
			{"names.$", nameChanger.NewName},
		}},
	}
	_,err := col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println(err)
		http.Error(w, "Did not upload", 418)
	}
	os.Rename("D:\\Data/"+nameChanger.Login+"/"+nameChanger.OldName+".mp3","D:\\Data/"+nameChanger.Login+"/"+nameChanger.NewName+".mp3")
	w.WriteHeader(200)
}

//
func audioHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.Replace(r.URL.Path, "/audio/", "", 1)
	if r.Method != http.MethodPost {
		//метод неправильный
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//проверка, что путь правильный
	if !(name == "addSong" || name == "getNamesOfSongs" || name=="changeName") {
		//путь неправильный
		http.Error(w, "Wrong path", 418)
		return
	}
	wasError, client := wasErrorInConnectToDb()
	if wasError {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if name=="changeName"{
		var nameChanger newName
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&nameChanger)
		if err != nil {
			http.Error(w, "Error in decoder", 418)
			return
		}
		if !checkFormValues(nameChanger.Login, nameChanger.Password, nameChanger.OldName, w) {
			if nameChanger.NewName ==""{
				http.Error(w,"Error in newName", 418)
				return
			}
			if nameChanger.OldName==nameChanger.NewName{
				http.Error(w,"OldName = NewName", 418)
				return
			}
			changeName(w,client,nameChanger)
		}
	}
	if name == "addSong" {
		var audio Audio
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&audio)
		if err != nil {
			http.Error(w, "Error in decoder", 418)
			return
		}
		if !checkFormValues(audio.Login, audio.Password, audio.Name, w) {
			uploadTable(w, client, audio)
		}
	}
	if name == "getNamesOfSongs" {
		var user User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			http.Error(w, "Error in decoder", 418)
			return
		}
		if !checkFieldsToAuth(user, w) {
			checker, names := isUserInTable(client, user)
			if !checker {
				http.Error(w, "Not exist", 418)
				return
			}
			files, err := ioutil.ReadDir("D:\\data/"+user.Login)
			if err != nil {
				http.Error(w, "Error in reading files", 418)
				return
			}
			var songs [][]byte
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), "mp3"){
					fmt.Println(f.Name())
					file, err:=ioutil.ReadFile("D:\\data/"+user.Login+"/"+f.Name())
					if err!=nil{
						http.Error(w, "Error in reading", 418)
						return
					}
					songs=append(songs, file)
				}
			}
			jsonFile, err := json.Marshal(Answer{
				Login:    user.Login,
				Names:    names,
				Mp3Files: songs,
			})
			if err != nil {
				http.Error(w, "Error in marshal", 418)
				return
			}
			w.WriteHeader(200)
			w.Write(jsonFile)
		}
	}
}

//проверяет, есть ли такой пользователь в таблице
func isUserInTable(client *mongo.Client, user User) (bool, []string) {
	col := client.Database(dbName).Collection(colName)
	filter := bson.D{{"login", user.Login},{"password",user.Password}}
	var result User
	err := col.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Printf("Not found user in db")
		return false, nil
	}else {
		log.Printf("Found a single document")
		return true, result.Names
	}
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("server [net/http] method [%s] connection from [%v]", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}
