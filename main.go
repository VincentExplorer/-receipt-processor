package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/google/uuid"
)

type Receipt struct {
    Retailer     string `json:"retailer"`
    PurchaseDate string `json:"purchaseDate"`
    PurchaseTime string `json:"purchaseTime"`
    Items        []Item `json:"items"`
    Total        string `json:"total"`
}

type Item struct {
    ShortDescription string `json:"shortDescription"`
    Price            string `json:"price"`
}

type ProcessReceiptResponse struct {
    ID string `json:"id"`
}

type PointsResponse struct {
    Points int `json:"points"`
}

var (
    receipts     = make(map[string]int)
    receiptsLock sync.Mutex
)

func main() {
    http.HandleFunc("/receipts/process", processReceiptHandler)
    http.HandleFunc("/receipts/", receiptsHandler)
    fmt.Println("Starting server on port 8080")
    http.ListenAndServe(":8080", nil)
}

func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var receipt Receipt
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&receipt)
    if err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    if !validateReceipt(receipt) {
        http.Error(w, "Invalid receipt data", http.StatusBadRequest)
        return
    }

    points := calculatePoints(receipt)
    id := uuid.New().String()

    receiptsLock.Lock()
    receipts[id] = points
    receiptsLock.Unlock()

    response := ProcessReceiptResponse{ID: id}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func receiptsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    path := r.URL.Path
    prefix := "/receipts/"
    if !strings.HasPrefix(path, prefix) {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    path = strings.TrimPrefix(path, prefix)
    parts := strings.Split(path, "/")
    if len(parts) != 2 || parts[1] != "points" {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    id := parts[0]

    receiptsLock.Lock()
    points, ok := receipts[id]
    receiptsLock.Unlock()
    if !ok {
        http.Error(w, "No receipt found for that id", http.StatusNotFound)
        return
    }

    response := PointsResponse{Points: points}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func validateReceipt(receipt Receipt) bool {
    if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) == 0 {
        return false
    }

    retailerPattern := `^[\w\s\-&]+$`
    matched, err := regexp.MatchString(retailerPattern, receipt.Retailer)
    if err != nil || !matched {
        return false
    }

    totalPattern := `^\d+\.\d{2}$`
    matched, err = regexp.MatchString(totalPattern, receipt.Total)
    if err != nil || !matched {
        return false
    }

    _, err = time.Parse("2006-01-02", receipt.PurchaseDate)
    if err != nil {
        return false
    }

    _, err = time.Parse("15:04", receipt.PurchaseTime)
    if err != nil {
        return false
    }

    for _, item := range receipt.Items {
        if item.ShortDescription == "" || item.Price == "" {
            return false
        }

        itemDescPattern := `^[\w\s\-]+$`
        matched, err = regexp.MatchString(itemDescPattern, item.ShortDescription)
        if err != nil || !matched {
            return false
        }

        matched, err = regexp.MatchString(totalPattern, item.Price)
        if err != nil || !matched {
            return false
        }
    }

    return true
}

func calculatePoints(receipt Receipt) int {
    points := 0

    retailerAlnum := regexp.MustCompile(`[a-zA-Z0-9]`)
    points += len(retailerAlnum.FindAllString(receipt.Retailer, -1))

    if isRoundDollarAmount(receipt.Total) {
        points += 50
    }

    if isMultipleOfQuarter(receipt.Total) {
        points += 25
    }

    points += (len(receipt.Items) / 2) * 5

    for _, item := range receipt.Items {
        desc := strings.TrimSpace(item.ShortDescription)
        if len(desc)%3 == 0 {
            price, _ := parseAmount(item.Price)
            itemPoints := int(price*0.2 + 0.999999)
            points += itemPoints
        }
    }

    purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
    day := purchaseDate.Day()
    if day%2 == 1 {
        points += 6
    }

    purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
    afterTime, _ := time.Parse("15:04", "14:00")
    beforeTime, _ := time.Parse("15:04", "16:00")
    if purchaseTime.After(afterTime) && purchaseTime.Before(beforeTime) {
        points += 10
    }

    return points
}

func isRoundDollarAmount(total string) bool {
    amount, err := parseAmount(total)
    if err != nil {
        return false
    }
    return amount == float64(int(amount))
}

func isMultipleOfQuarter(total string) bool {
    amount, err := parseAmount(total)
    if err != nil {
        return false
    }
    remainder := int(amount*100) % 25
    return remainder == 0
}

func parseAmount(amountStr string) (float64, error) {
    amount, err := strconv.ParseFloat(amountStr, 64)
    return amount, err
}
