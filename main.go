package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName   string = "localhost:27017"
	dbName     string = "demo_todo"
	collection string = "todo"
	port       string = ":9000"
)

type (
	TodoModel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Completed bool          `bson:"completed"`
		CreatedAt time.Time     `bson:"created_at"`
	}

	Todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {
	rnd = renderer.New()
	sess, err := mgo.Dial(hostName)
	if err != nil {
		log.Fatal(err)
	}
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)

}
func main() {

	fmt.Println("Hello World")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
		if err != nil {
			log.Fatal(err)
		}
	})
	r.Mount("/todo", todoHandler())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Print("Starting server at port", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("errror serving server ", err)
	}

}

func todoHandler() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})

	return rg

}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	todos := []TodoModel{}
	if err := db.C(collection).Find(bson.M{}).All(&todos); err != nil {
		fmt.Printf("errror fetching todos")
		return
	}
	todosList := []Todo{}
	for _, t := range todos {
		todosList = append(todosList, Todo{
			ID:        t.ID.Hex(),
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todosList,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	var todoModel TodoModel
	json.NewDecoder(r.Body).Decode(&todo)
	if todo.Title == "" {
		json.NewEncoder(w).Encode("title missing")
		return
	}

	todoModel = TodoModel{
		ID:        bson.NewObjectId(),
		Title:     todo.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	if err := db.C(collection).Insert(&todoModel); err != nil {
		fmt.Println(err)
	}
	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "todo created successfully",
		"todo_id": todoModel.ID.Hex(),
	})
}
func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if !bson.IsObjectIdHex(id) {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}
	var todo Todo

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "error udpatign todo",
		})
	}
	if todo.Title == "" {

		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "title field requred",
		})

		return
	}

	if err := db.C(collection).Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{
		"title": todo.Title, "completed": todo.Completed,
	}); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to update todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "successfully updated id",
	})
}
func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if !bson.IsObjectIdHex(id) {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	if err := db.C(collection).RemoveId(bson.ObjectIdHex(id)); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to delete todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "successfully deleted id",
	})
}
