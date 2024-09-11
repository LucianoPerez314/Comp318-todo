package todo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type toDo struct {
	Id          int
	Description string
}

type toDoList struct {
	mtx    sync.Mutex
	list   map[int]string
	NextId int
}

type description struct {
	Description string
}

func (t *toDoList) getAll(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//Lock mtx
	t.mtx.Lock()
	defer t.mtx.Unlock()
	//Make slice of Todos from todoList
	toDos := make([]toDo, 0, len(t.list))
	for id, desc := range t.list {
		toDos = append(toDos, toDo{Id: id, Description: desc})
	}
	//Marshal go toDo struct into json object.
	encoded, err := json.Marshal(toDos)
	if err != nil {
		msg := fmt.Sprintf(`"todos corrupted: %s"`, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	//Respond with json object.
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (t *toDoList) createToDo(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//Lock mtx
	t.mtx.Lock()
	defer t.mtx.Unlock()
	//Get description of todo to add.
	jsonDesc, err := io.ReadAll(r.Body)
	//Unmarshal into a description
	var desc description
	json.Unmarshal(jsonDesc, &desc)
	//Append todo
	t.list[t.NextId] = desc.Description
	//Marshal go toDo struct into json object.
	encoded, err := json.Marshal(toDo{Description: desc.Description, Id: t.NextId})
	t.NextId = t.NextId + 1
	if err != nil {
		msg := fmt.Sprintf(`"todos corrupted: %s"`, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	//Respond with json object.
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (t *toDoList) retrieveToDo(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//Lock mtx
	t.mtx.Lock()
	defer t.mtx.Unlock()
	//Get id of todo.
	// Split the path (e.g., /todos/123)
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	//Get description from id.
	desc := t.list[id]
	//Marshal go toDo struct into json object.
	encoded, err := json.Marshal(toDo{Description: desc, Id: id})
	if err != nil {
		msg := fmt.Sprintf(`"todos corrupted: %s"`, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	//Respond with json object.
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (t *toDoList) createOrReplaceToDo(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//Lock mtx
	t.mtx.Lock()
	defer t.mtx.Unlock()
	//Get description of todo to add or replace
	jsonDesc, err := io.ReadAll(r.Body)
	//Unmarshal into a description
	var desc description
	json.Unmarshal(jsonDesc, &desc)
	// Split the path (e.g., /todos/123) for id
	parts := strings.Split(r.URL.Path, "/")
	//Create Todo
	var id int
	var erru error
	if parts[2] == "" {
		id = t.NextId
		t.NextId = t.NextId + 1
	} else { //Replace Todo
		id, erru = strconv.Atoi(parts[2])
		if erru != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
	}
	t.list[id] = desc.Description
	encoded, err := json.Marshal(toDo{Description: desc.Description, Id: id})
	if err != nil {
		msg := fmt.Sprintf(`"todos corrupted: %s"`, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	//Respond with json object.
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (t *toDoList) deleteToDo(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//Lock mtx
	t.mtx.Lock()
	defer t.mtx.Unlock()
	//Get id of todo.
	// Split the path (e.g., /todos/123)
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	//Delete ToDo
	desc := t.list[id]
	delete(t.list, id)
	//Marshal go toDo struct into json object.
	encoded, err := json.Marshal(toDo{Description: desc, Id: id})
	if err != nil {
		msg := fmt.Sprintf(`"todos corrupted: %s"`, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	//Respond with json object.
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (t *toDoList) optionsAllToDos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Allow", "GET,POST")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func New() http.Handler {
	todos := toDoList{mtx: sync.Mutex{}, list: make(map[int]string)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", todos.getAll)
	mux.HandleFunc("POST /todos", todos.createToDo)
	mux.HandleFunc("GET /todos/", todos.retrieveToDo)
	mux.HandleFunc("PUT /todos/", todos.createOrReplaceToDo)
	mux.HandleFunc("DELETE /todos/", todos.deleteToDo)
	mux.HandleFunc("OPTIONS /todos", todos.optionsAllToDos)
	mux.HandleFunc("OPTIONS /todos/", todos.optionsAllToDos)
	return mux
}
