package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/go-ping/ping"
)

// Структура для хранения данных о результате пинга
type PingResult struct {
    IP          string        `json:"ip"`           // IP-адрес контейнера
    Duration    time.Duration `json:"duration"`     // Длительность пинга
    SuccessDate *time.Time    `json:"success_date"` // Время успешной отправки (может быть nil)
}

// Функция для выполнения пинга контейнера
func pingContainer(host string) (bool, string, time.Duration, error) {
    pinger, err := ping.NewPinger(host)
    if err != nil {
        return false, "", 0, fmt.Errorf("failed to create pinger: %w", err)
    }

    // Устанавливаем привилегии для отправки ICMP-пакетов (требуется root)
    pinger.SetPrivileged(true)

    // Задаем таймаут для ожидания ответа
    pinger.Timeout = 1 * time.Second

    // Запускаем пинг
    err = pinger.Run()
    if err != nil {
        return false, "", 0, fmt.Errorf("failed to run pinger: %w", err)
    }

    // Получаем статистику
    stats := pinger.Statistics()

    // Проверяем, был ли хотя бы один успешный ответ
    if stats.PacketsRecv > 0 {
        return true, stats.IPAddr.String(), stats.AvgRtt, nil
    }

    return false, "", 0, nil
}

// Функция для отправки данных на API
func sendPingResult(apiURL string, result PingResult) error {
    // Сериализуем данные в JSON
    jsonData, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal ping result: %w", err)
    }

    // Создаем HTTP-запрос
    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Добавляем API-ключ в заголовок
    var apiKey = os.Getenv("API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable is not set")
    }
    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    // Отправляем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send POST request: %w", err)
    }
    defer resp.Body.Close()

    // Проверяем статус ответа
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}

func main() {
    // Проверяем, запущены ли мы от имени root
    if os.Geteuid() != 0 {
        log.Fatal("This program must be run as root to send ICMP packets.")
    }

    // URL API для отправки данных
    apiURL := "http://my-api:8080/ping" // Замените на реальный URL вашего API

    // Контейнеры для пинга
    containers := []string{
        "sleeper1",
        "sleeper2",
        "sleeper3",
        "sleeper4",
        "sleeper5",
    }

    // Бесконечный цикл пинга
    for {
        for _, container := range containers {
            isReachable, ip, duration, err := pingContainer(container)
            if err != nil {
                log.Printf("Error pinging %s: %v\n", container, err)
                continue
            }

            // Создаем объект с результатами пинга
            result := PingResult{
                IP:       ip,
                Duration: duration,
            }

            // Если пинг не был успешным, длительность будет 0
            if isReachable {
                successDate := time.Now()
                result.SuccessDate = &successDate
            }

            // Отправляем данные через API
            if err := sendPingResult(apiURL, result); err != nil {
                log.Printf("Failed to send ping result for %s: %v\n", container, err)
            } else {
                log.Printf("Ping result for %s sent successfully\n", container)
            }
        }

        // Пауза между пингами
        time.Sleep(10 * time.Second)
    }
}