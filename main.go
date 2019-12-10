package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type JsonMessage struct {
	Time int `json:"time"`
	LiftID int `json:"liftID"`
}


var DB *sql.DB
var dataBase = "admin:criminal51@tcp(34.83.152.203:3306)" +
	"/skiDatabase?loc=Local&parseTime=true"
//var dataBase = "root:criminal51@tcp(127.0.0.1:3306)/skiDatabase?loc=Local&parseTime=true"

func initDatabase(){
	var err error
	DB, err = sql.Open("mysql", dataBase)
	if err != nil{
		log.Fatalln("open database failed:", err)
	}

	DB.SetMaxOpenConns(20000)
	DB.SetMaxIdleConns(1000)
	DB.SetConnMaxLifetime(time.Second * 8)

	DB.Ping()
	fmt.Println("Database init")
}

func skierData(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	skierId := vars["skierId"]
	seasonId := vars["seasonId"]
	dayId := vars["dayId"]
	resortId := vars["resortId"]

	if r.Method == "GET" {
	statement := "SELECT SUM(vertical) as total FROM skiRecords WHERE skiRecords.skierId = ? AND skiRecords.dayId = ?"
	stmtQuery, err := DB.Prepare(statement)
	if err != nil{
		log.Fatalln(err)
	}

	rows, err := stmtQuery.Query(skierId, dayId)

	for rows.Next(){
		var total int
		if err := rows.Scan(&total); err != nil{
			log.Fatalln(err)
		}
	}
	if err := rows.Err(); err != nil{
		log.Fatalln(err)
	}
	rows.Close()
	stmtQuery.Close()

	} else if r.Method == "POST" {
		statment := "INSERT INTO skiRecords (recordId, skierId, resortId, season, dayId, skiTime, LiftId, vertical) " +
			"VALUES (?,?,?,?,?,?,?,?)"
		stmtIns, err := DB.Prepare(statment)
		if err != nil{
			panic(err.Error())
		}
		recordId := uuid.New()
		var p JsonMessage
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil{
			http.Error(w, err.Error(), 500)
			return
		}
		skiTime := p.Time
		LiftId := p.LiftID
		vertical := p.LiftID * 10
		_, err = stmtIns.Exec(recordId, skierId, resortId, seasonId, dayId, skiTime, LiftId, vertical)
		if err != nil{
			panic(err.Error())
		}
		stmtIns.Close()
	}
}

func Hello(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "Hello, World!")
}

func main() {
	initDatabase()
	rtr := mux.NewRouter()
	rtr.HandleFunc("/skiers/{resortId:[0-9]+}/seasons/{seasonId:[0-9]+}/days/{dayId:[0-9]+}/skiers/{skierId:[0-9]+}", skierData)
	//http.HandleFunc("/skiers/{resortId:[0-9]+}/seasons/{seasonId:[0-9]+}/days/{dayId:[0-9]+}/skiers/{skierId:[0-9]+}", skierData)
	//http.HandleFunc("/", Hello)
	rtr.HandleFunc("/", Hello)
	http.Handle("/", rtr)
	fmt.Println("Server starts...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error", err)
	}

}