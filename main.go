/*
Вариант 2
Нужно реализовать сервис для учета времени сотрудников на рабочем месте.
ID, FIO, DEPARTMENT, POSITION

In(Пришел)  -> Datetime
Out(Ушел)   -> Datetime

GET EMPLOYEE
POST EMPLOYEE

GET ALLTIME(day or month) выдать сколько времени провел на рабочем месте за указанный период времени.
*/
package main

import (
	"database/sql"
	"emplTime/dbutils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
	_ "github.com/mattn/go-sqlite3"
)

// DB Driver visible to whole program
var DB *sql.DB

// EmployeResource is the model for holding employe information
type EmployeResource struct {
	ID         int
	FIO        string
	Department string
	Position   string
}

// TimeResource holds arrival and leaving information
type TimeResource struct {
	ID        int
	In        string
	Out       string
	EmployeID int
}

// Register adds paths and routes to container for EmployeResource
func (t *EmployeResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/v1/employe").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{employe-id}").To(t.getEmploye))
	ws.Route(ws.POST("").To(t.createEmploye))
	ws.Route(ws.DELETE("/{employe-id}").To(t.removeEmploye))

	container.Add(ws)
}

// Register adds paths and routes to container for TimeResource
func (t *TimeResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/v1/event").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{employe-id}").To(t.getEventsForEmployeID))
	ws.Route(ws.POST("").To(t.createEvent))
	ws.Route(ws.DELETE("/{event-id}").To(t.removeEvent))

	container.Add(ws)
}

// GET http://localhost:8000/v1/employe/1
func (t EmployeResource) getEmploye(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("employe-id")
	err := DB.QueryRow("select ID, FIO, DEPARTMENT,POSITION FROM employe where id=?", id).Scan(&t.ID, &t.FIO, &t.Department, &t.Position)
	if err != nil {
		log.Println(err)
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Train could not be found.")
	} else {
		response.WriteEntity(t)
	}
}

func (t TimeResource) getSlieceEventsByEmpID(eid string) ([]TimeResource, error) {
	employe_id, _ := strconv.Atoi(eid)
	// Конвертируем время типа INTEGER в виде Unix.timestampt и возвращаем в виде "2006-01-02 08:00:05"
	rows, err := DB.Query("select ID, datetime(ARRIVAL_TIME,'unixepoch'), datetime(LEAVING_TIME,'unixepoch'), EMPLOYE_ID FROM events where EMPLOYE_ID=?", employe_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]TimeResource, 0)
	var eventIs TimeResource
	for rows.Next() {
		err = rows.Scan(&eventIs.ID, &eventIs.In, &eventIs.Out, &eventIs.EmployeID)
		if err != nil {
			return nil, err
		}
		log.Println(&eventIs.ID, &eventIs.In, &eventIs.Out, &eventIs.EmployeID)
		events = append(events, eventIs)
	}
	return events, nil
}

// GET http://localhost:8000/v1/event/1
func (t TimeResource) getEventsForEmployeID(request *restful.Request, response *restful.Response) {
	employe_id := request.PathParameter("employe-id")
	events, err := t.getSlieceEventsByEmpID(employe_id)
	if err != nil {
		log.Println(err)
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Employe could not be found.")
	} else {
		response.WriteEntity(events)
	}
}

// POST http://localhost:8000/v1/employe/1
func (t EmployeResource) createEmploye(request *restful.Request, response *restful.Response) {
	log.Println(request.Request.Body)
	decoder := json.NewDecoder(request.Request.Body)
	var b EmployeResource
	err := decoder.Decode(&b)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(b.ID, b.FIO, b.Department, b.Position)

	// Error handling is obvious here. So omitting...
	statement, _ := DB.Prepare("insert into employe (ID, FIO, DEPARTMENT, POSITION) values (?, ?, ?, ?)")
	result, err := statement.Exec(b.ID, b.FIO, b.Department, b.Position)
	if err == nil {
		newID, _ := result.LastInsertId()
		b.ID = int(newID)
		response.WriteHeaderAndEntity(http.StatusCreated, b)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}

// POST http://localhost:8000/v1/event/1
func (t TimeResource) createEvent(request *restful.Request, response *restful.Response) {
	log.Println(request.Request.Body)
	decoder := json.NewDecoder(request.Request.Body)
	var b TimeResource
	err := decoder.Decode(&b)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(b.ID, func(bin string) time.Time { time, _ := time.Parse(time.RFC3339, bin); return time }(b.In).Unix(), func(bin string) time.Time { time, _ := time.Parse(time.RFC3339, bin); return time }(b.Out).Unix(), b.EmployeID)

	// Error handling is obvious here. So omitting...
	statement, _ := DB.Prepare("insert into events (ID, ARRIVAL_TIME, LEAVING_TIME, EMPLOYE_ID) values (?, ?, ?, ?)")
	//Парсим  строки со временем типа "2006-01-02T08:00:05Z" в тип time.Time и затем конвертируем в Unix.time перед добавлением в базу
	result, err := statement.Exec(b.ID, func(bin string) time.Time { time, _ := time.Parse(time.RFC3339, bin); return time }(b.In).Unix(), func(bin string) time.Time { time, _ := time.Parse(time.RFC3339, bin); return time }(b.Out).Unix(), b.EmployeID)
	if err == nil {
		newID, _ := result.LastInsertId()
		b.ID = int(newID)
		response.WriteHeaderAndEntity(http.StatusCreated, b)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}

// DELETE http://localhost:8000/v1/employe/1
func (t EmployeResource) removeEmploye(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("employe-id")
	statement, _ := DB.Prepare("delete from employe where id=?")
	_, err := statement.Exec(id)
	if err == nil {
		response.WriteHeader(http.StatusOK)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}

// DELETE http://localhost:8000/v1/event/1
func (t TimeResource) removeEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	statement, _ := DB.Prepare("delete from events where id=?")
	_, err := statement.Exec(id)
	if err == nil {
		response.WriteHeader(http.StatusOK)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}

func durateEvent(t1, t2 time.Time) time.Duration {
	return t2.Sub(t1)
}

func main() {
	fmt.Println(time.Now())
	tmz1, _ := time.Parse(time.RFC3339, "2006-01-02T08:00:05Z")
	tmz2, _ := time.Parse(time.RFC3339, "2006-01-02T18:00:05Z")
	fmt.Println(tmz1.Unix())
	fmt.Println(tmz2.Unix())
	fmt.Println(durateEvent(tmz1, tmz2))

	var err error
	DB, err = sql.Open("sqlite3", "./employes.db")
	if err != nil {
		log.Println("Driver creation failed!")
	}
	dbutils.Initialize(DB)
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	t := EmployeResource{}
	t.Register(wsContainer)

	e := TimeResource{}
	e.Register(wsContainer)

	log.Printf("start listening on localhost:8000")
	server := &http.Server{Addr: ":8000", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
