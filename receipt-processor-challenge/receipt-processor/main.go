package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid" 
	"math"                   
	"net/http"              
	"regexp"                 
	"strconv"                
	"strings"               
	"sync"                   
	"time"                  
)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// points response
type PointsResponse struct {
	Points int `json:"points"`
}

// receiptID response
type IDResponse struct {
	ID string `json:"id"`
}

var (
	receipts = make(map[string]Receipt) // Key: receiptID, Val: Receipt
	points   = make(map[string]int)     // Key: receiptID, Val: total points
	mu       sync.Mutex                 
)

// helper function to calculate total points for a receipt
func calculatePoints(receipt Receipt) int {
	points := 0

	// 1 point for every alphanumeric character in the retailer name
	alphaNum := regexp.MustCompile(`[a-zA-Z0-9]`)
	points += len(alphaNum.FindAllString(receipt.Retailer, -1))

	// 50 points if the total is a round dollar amount
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err == nil && total == float64(int(total)) {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25
	if err == nil && math.Mod(total, 0.25) == 0 {
		points += 25
	}

	// 5 points for every two items
	points += (len(receipt.Items) / 2) * 5

	// points for item description length and price
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc) %3 == 0 {
			itemPrice, err := strconv.ParseFloat(item.Price, 64)
			if err == nil {
				points += int(math.Ceil(itemPrice * 0.2))
			}
		}
	}

	// 6 points if the day in purchase date is odd
	parsedDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err == nil && parsedDate.Day()%2 != 0 {
		points += 6
	}

	// 10 points if the purchase time is between 2:00pm and 4:00pm
	parsedTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err == nil && parsedTime.Hour() == 14 {
		points += 10
	}

	return points
}

// POST /receipts/process endpoint
func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt

	// decode the JSON body into a Receipt struct
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Invalid receipt", http.StatusBadRequest)
		return
	}

	// generate unique ID for the receipt
	receiptID := uuid.New().String()

	// store the receipt and its calculated points
	mu.Lock()
	receipts[receiptID] = receipt
	points[receiptID] = calculatePoints(receipt)
	mu.Unlock()

	// respond with unique receipt ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IDResponse{ID: receiptID})
}

// GET /receipts/{id}/points endpoint
func getPointsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the receipt ID from the URL path
	receiptID := strings.TrimPrefix(r.URL.Path, "/receipts/")
	if strings.HasSuffix(receiptID, "/points") {
		receiptID = strings.TrimSuffix(receiptID, "/points")
	}

	// retrieve points for given receiptID
	mu.Lock()
	receiptPoints, exists := points[receiptID]
	mu.Unlock()

	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	// respond with the points as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PointsResponse{Points: receiptPoints})
}

// set up and start the HTTP server
func main() {
	//POST
	http.HandleFunc("/receipts/process", processReceiptHandler) 
	//GET 
	http.HandleFunc("/receipts/", getPointsHandler)             

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}