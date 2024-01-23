package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	. "todo/models"
	"todo/templates"
)

func main() {
	// Create sqlx sqlite3 connection
	db, err := sqlx.Open("sqlite3", "todos.db")
	if err != nil {
		logrus.WithError(err).Fatal("error opening sqlite3 database")
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", Index(db))
	router.Post("/todo", CreateTodo(db))
	router.Get("/todo", GetTodos(db))
	router.Put("/todo/complete", ToggleComplete(db))
	router.Delete("/todo", DeleteTodo(db))

	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logrus.Fatal(server.ListenAndServe())
}

func Index(db *sqlx.DB) http.HandlerFunc {
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
			err := templates.ErrorMessage(Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}).Render(r.Context(), w)
			if err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		if err := templates.Index(todos).Render(r.Context(), w); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func CreateTodo(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PostFormValue("name")
		if name == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			err := templates.ErrorMessage(Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "name is required",
			}).Render(r.Context(), w)
			if err != nil {
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
			err := templates.ErrorMessage(Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}).Render(r.Context(), w)
			if err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		err = templates.TodoItem(Todo{
			ID:        id,
			Name:      name,
			Completed: false,
			CreatedAt: now,
			UpdatedAt: now,
		}).Render(r.Context(), w)
		if err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func GetTodos(db *sqlx.DB) http.HandlerFunc {
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
			err := templates.ErrorMessage(Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}).Render(r.Context(), w)
			if err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		if err := templates.TodoList(todos).Render(r.Context(), w); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func ToggleComplete(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			err := templates.ErrorMessage(Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "id is required",
			}).Render(r.Context(), w)
			if err != nil {
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
			err := templates.ErrorMessage(Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}).Render(r.Context(), w)
			if err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := templates.TodoItem(todo).Render(r.Context(), w); err != nil {
			logrus.WithError(err).Error("error executing template")
		}
	}
}

func DeleteTodo(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.Header().Set("HX-Retarget", "#todo-list")
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusOK)
			err := templates.ErrorMessage(Error{
				StatusCode: 400,
				Status:     "Bad Request",
				Message:    "id is required",
			}).Render(r.Context(), w)
			if err != nil {
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
			err := templates.ErrorMessage(Error{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Message:    err.Error(),
			}).Render(r.Context(), w)
			if err != nil {
				logrus.WithError(err).Error("error executing template")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
