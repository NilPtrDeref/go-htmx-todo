package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"time"
)

type Todo struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Completed bool      `db:"completed"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Error struct {
	StatusCode int
	Status     string
	Message    string
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// Create sqlx sqlite3 connection
	db, err := sqlx.Open("sqlite3", "todos.db")
	if err != nil {
		logrus.WithError(err).Fatal("error opening sqlite3 database")
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", Index(tmpl, db))
	router.Post("/todo", CreateTodo(tmpl, db))
	router.Get("/todo", GetTodos(tmpl, db))
	router.Put("/todo/complete", ToggleComplete(tmpl, db))
	router.Delete("/todo", DeleteTodo(tmpl, db))

	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logrus.Fatal(server.ListenAndServe())
}

func Index(tmpl *template.Template, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var todos []Todo
		err := db.Select(&todos, `
		SELECT 
			id, name, completed, created_at, updated_at
		FROM 
			todos`)
		if err != nil {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		if err := tmpl.ExecuteTemplate(w, "index", todos); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func CreateTodo(tmpl *template.Template, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PostFormValue("name")
		if name == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "name is required",
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		now := time.Now()
		var id int
		err := db.Get(&id, `
		INSERT INTO 
			todos(name, completed, created_at, updated_at)
		VALUES(?, ?, ?, ?)
		RETURNING id`,
			name, false, now, now)
		if err != nil {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := tmpl.ExecuteTemplate(w, "todo", Todo{
			ID:        id,
			Name:      name,
			Completed: false,
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func GetTodos(tmpl *template.Template, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var todos []Todo
		err := db.Select(&todos, `
		SELECT 
			id, name, completed, created_at, updated_at
		FROM 
			todos`)
		if err != nil {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		if err := tmpl.ExecuteTemplate(w, "todo-list", todos); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func ToggleComplete(tmpl *template.Template, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "id is required",
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		var todo Todo
		err := db.Get(&todo, `
		UPDATE todos
		SET completed = NOT completed, updated_at = ?
		WHERE id = ?
		RETURNING id, name, completed, created_at, updated_at`, time.Now(), id)
		if err != nil {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := tmpl.ExecuteTemplate(w, "todo", todo); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func DeleteTodo(tmpl *template.Template, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "id is required",
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		_, err := db.Exec(`
		DELETE FROM todos
		WHERE id = ?`, id)
		if err != nil {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			if err := tmpl.ExecuteTemplate(w, "error", Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}); err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
