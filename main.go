

package main

import (
	_ "crypto/dsa"
	"database/sql"
	_ "database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"log"
	_ "log"
	_ "math/rand"
	"net/http"
	_ "net/http"
	"time"
	_ "time"
)
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
type authCredentials struct {
	Username string `json:"username"`
	Password  string `json:"password"`
}
var jwtKey = []byte("my_secret_key")

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "your-password"
	dbname   = "manidharmulagapaka"
)
type person struct {
	Username     string
	Password     string
}
var clients = make(map[*websocket.Conn]string)

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
var db, err = sql.Open("postgres",psqlInfo)


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
	return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var connectionUser authCredentials
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewDecoder(r.Body).Decode(&connectionUser); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	clients[ws] = connectionUser.Username
}
func login(w http.ResponseWriter, r *http.Request) {
	var parameters authCredentials
	if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	var sqlStatement = `SELECT username FROM users WHERE username=$1 AND password=$2`
	var result_username string
	var row = db.QueryRow(sqlStatement, parameters.Username, parameters.Password)
	switch err:= row.Scan(&result_username); err{
	case sql.ErrNoRows:
		fmt.Println("No Rows were returned")
		w.WriteHeader(404)
	case nil:
		fmt.Println(result_username)
	default:
		panic(err)
	}
	defer r.Body.Close()
}
func signup(w http.ResponseWriter, r *http.Request){
	var signupCredentials authCredentials
	if err := json.NewDecoder(r.Body).Decode(&signupCredentials); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	var username = signupCredentials.Username
	var password = signupCredentials.Password

	var sqlStatement = `INSERT INTO users (username,password)
	VALUES ($1,$2)`
	_, err = db.Exec(sqlStatement, username,password);
	if err != nil {
		panic(err)
	}
	person := person{username,password}
	jsonResponse,jsonError := json.Marshal(person)
	if(jsonError != nil){
		fmt.Println("unable to read json")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: signupCredentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, _ := token.SignedString(jwtKey)

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.Write([]byte(tokenString))
}

func homepage(w http.ResponseWriter,r *http.Request){

}
func main() {

	router := mux.NewRouter()

	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.HandleFunc("/homepage",homepage).Methods("POST")
	defer db.Close()


	//sqlStatement := `CREATE TABLE users (
	//										username TEXT,
	//										password TEXT
	//									)`
	//_, err = db.Exec(sqlStatement)

	if err != nil {
		fmt.Print(err)
	}
	log.Fatal(http.ListenAndServe(":8847", router))
}