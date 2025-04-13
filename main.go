package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    _ "github.com/lib/pq"
    "os"
)

type Todo struct {
    ID   int
    Task string
}

var db *sql.DB

func main() {
    // Підключення до PostgreSQL
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), 
        os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
    
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Перевірка підключення
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    // Створення таблиці
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
        id SERIAL PRIMARY KEY,
        task TEXT NOT NULL
    )`)
    if err != nil {
        log.Fatal(err)
    }

    // Налаштування маршрутів
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/add", addHandler)

    log.Println("Server starting on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, task FROM todos")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var todos []Todo
    for rows.Next() {
        var t Todo
        err := rows.Scan(&t.ID, &t.Task)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        todos = append(todos, t)
    }

    fmt.Fprintf(w, `<html>
        <body>
            <h1>TODO List</h1>
            <form action="/add" method="post">
                <input type="text" name="task">
                <input type="submit" value="Add">
            </form>
            <ul>`)
    for _, t := range todos {
        fmt.Fprintf(w, "<li>%d: %s</li>", t.ID, t.Task)
    }
    fmt.Fprintf(w, `</ul></body></html>`)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    task := r.FormValue("task")
    if task == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    _, err := db.Exec("INSERT INTO todos (task) VALUES ($1)", task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}