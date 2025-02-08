package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "ping_api/api/internal/config"
    _ "github.com/lib/pq"
    "time"
    "os"
)

var db *sql.DB

// API-ключ (должен быть защищён и храниться в безопасном месте, например, в переменных окружения)
var apiKey = os.Getenv("API_KEY")

type PingResult struct {
    ID          int            `json:"id"`
    IP          string         `json:"ip"`
    SuccessDate *time.Time     `json:"success_date"`
    Duration    time.Duration  `json:"duration"`
}

// Middleware для проверки API-ключа
func apiKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Получаем API-ключ из заголовка
        key := r.Header.Get("X-API-Key")
        if key != apiKey {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}

func addPingResultHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(&w)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var result PingResult
    if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
        http.Error(w, "Failed to decode request body", http.StatusBadRequest)
        return
    }

    query := `
        INSERT INTO users (ip, success_date, duration)
        VALUES ($1, $2, $3)
    `
    _, err := db.Exec(query, result.IP, result.SuccessDate, result.Duration)
    if err != nil {
        http.Error(w, "Failed to save ping result", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Ping result saved successfully"))
}

func enableCORS(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func initDB(cfg *config.Config) error {
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
    )

    var err error
    for i := 0; i < 10; i++ {
        db, err = sql.Open("postgres", connStr)
        if err == nil {
            err = db.Ping()
            if err == nil {
                log.Println("Connected to the database")
                return nil
            }
        }
        log.Printf("Attempt %d: failed to connect to database, retrying...\n", i+1)
        time.Sleep(5 * time.Second)
    }
    return fmt.Errorf("failed to connect to database after multiple attempts: %v", err)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(&w)

    rows, err := db.Query("SELECT id, ip, success_date, duration FROM users")
    if err != nil {
        http.Error(w, "Failed to query database", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var results []PingResult
    for rows.Next() {
        var user PingResult
        if err := rows.Scan(&user.ID, &user.IP, &user.SuccessDate, &user.Duration); err != nil {
            log.Printf("Failed to scan row: %v", err)
            http.Error(w, "Failed to scan row", http.StatusInternalServerError)
            return
        }
        results = append(results, user)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, "Error during rows iteration", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(results); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

func main() {
    cfg := config.LoadConfig()

    if err := initDB(cfg); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Защищаем эндпоинт /ping с помощью API-ключа
    http.HandleFunc("/ping", apiKeyMiddleware(addPingResultHandler))

    // Эндпоинт /users остаётся открытым
    http.HandleFunc("/users", pingHandler)

    addr := cfg.Host + ":" + cfg.Port
    log.Printf("Starting server on %s\n", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Could not start server: %v\n", err)
    }
}