package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const dbName string = "DataBase"
const colName string = "User"

type User struct {
	Login string `json:"loginOfUser"`
	Password string `json:"passwordOfUser"`
	Names [] string `json:"namesOfSongs"`
}

type Answer struct{
	Error string `json:"error"`
	Names []string `json:"names"`
}

func main() {
	handler := http.NewServeMux()
	handler.HandleFunc("/getInApp/", Logger(authHandler))
	handler.HandleFunc("/audio", Logger(audioHandler))
	s := http.Server{
		Addr: ":8080",
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.Replace(r.URL.Path, "/getInApp/", "", 1)
	if r.Method == http.MethodPost {
		if name == "registration" || name == "enter" {
			var user User
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				if !checkFieldsToAuth(user, w) {
					wasError, client := wasErrorInConnectToDb()
					if wasError {
						w.WriteHeader(http.StatusNotFound)
					} else {
						if name == "registration" {
							if isUserInTable(w, r, client, user) {
								http.Error(w, "Already exist", 418)
							} else {
								addToTable(client, user)
								w.WriteHeader(http.StatusOK)
							}
						}
						if name == "enter" {
							if isUserInTable(w, r, client, user) {
								w.WriteHeader(http.StatusOK)
							} else {
								http.Error(w, "Not exist", 418)
							}
						}
					}
				}
			}
		} else {
			http.Error(w, "Wrong path", 418)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//проверка значений логина и пароля
func checkFieldsToAuth(user User, w http.ResponseWriter) bool {
	if user.Login == "" {
		http.Error(w, "Wrong login", 418)
		return true
	} else if user.Password == "" {
		http.Error(w, "Wrong password", 418)
		return true
	} else if user.Names==nil{
		http.Error(w, "Wrong names - nil", 418)
		return true
	} else if len(user.Names)!=0{
		http.Error(w, "Wrong names "+string(len(user.Names)), 418)
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
	log.Printf("Connected to MongoDB!")
	return wasError, client
}

//добавление нового пользоваателя в таблицу
func addToTable(client *mongo.Client, user User) {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := client.Database(dbName).Collection(colName)
	result, insertErr := col.InsertOne(ctx, user)
	if insertErr != nil {
		log.Printf("InsertOne ERROR:%v", insertErr)
		os.Exit(1)
	} else {
		newID := result.InsertedID
		fmt.Println("InsertedOne(), newID", newID)
		fmt.Println("InsertedOne(), newID type:", reflect.TypeOf(newID))
	}
}

func uploadTable(client *mongo.Client, login string, password string, name string) (bool, [] string){
	col := client.Database(dbName).Collection(colName)
	filter := bson.D{{"login", login}, {"password",password}}
	var result User
	err := col.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Println("Not found")
		return false, nil
	}else {
		fmt.Printf("Found a single document: %+v\n", result)
		checker := false
		for _, elem := range result.Names {
			if elem == name {
				checker = true
			}
		}
		if !checker {
			update := bson.D{
				{"$push", bson.D{
					{"names", name},
				}},
			}
			_, err := col.UpdateOne(context.TODO(), filter, update)
			fmt.Println(err)
			return true, append(result.Names, name)
		}else{
			fmt.Println("The same name")
			return false, nil
		}
	}
}

//
func audioHandler(w http.ResponseWriter, r *http.Request) {
	//name := strings.Replace(r.URL.Path, "/audio/", "", 1)
	wasError, client := wasErrorInConnectToDb()
	if wasError {
		w.WriteHeader(http.StatusNotFound)
	} else {
		login := r.PostFormValue("login")
		fmt.Println(login)
		password := r.PostFormValue("password")
		fmt.Println(password)
		nameOfSong := r.PostFormValue("nameOfSong")
		fmt.Println(nameOfSong)
		file, header, err := r.FormFile("audioFile")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer file.Close()
		checker,names:=uploadTable(client, login, password, nameOfSong)
		if(checker) {
			out, err := os.Create("D:\\newmus.mp3")
			if err != nil {
				fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
				return
			}
			defer out.Close()
			_, err = io.Copy(out, file)
			if err != nil {
				fmt.Println(w, err)
			}
			json_data2, err := json.Marshal(Answer{
				Error: "File uploaded successfully : "+header.Filename,
				Names: names,
			})

			if err != nil {

				log.Fatal(err)
			}
			w.Write(json_data2)
		}
	}
}

//проверяет, есть ли такой пользователь в таблице
func isUserInTable(w http.ResponseWriter, r *http.Request, client *mongo.Client, user User) bool {
	col := client.Database(dbName).Collection(colName)
	filter := bson.D{{"login", user.Login},{"password",user.Password}}
	var result User
	err := col.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Println("Not found")
		return false
	}else {
		fmt.Printf("Found a single document: %+v\n", result)
		return true
	}
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("server [net/http] method [%s] connection from [%v]", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}
