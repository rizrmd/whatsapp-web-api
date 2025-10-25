package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "github.com/lib/pq"
)

var (
	client     *whatsmeow.Client
	qrChannel  chan string
	webhookURL string
	isPaired   bool = false
)

// Response structures for API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SendRequest struct {
	Number  string `json:"number"`
	Message string `json:"message"`
}

type WebhookPayload struct {
	Event   string    `json:"event"`
	Message string    `json:"message,omitempty"`
	Sender  string    `json:"sender,omitempty"`
	Chat    string    `json:"chat,omitempty"`
	Time    time.Time `json:"time"`
}

func getDatabaseURL() string {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Please set it in .env file or as environment variable")
	}

	// Parse the URL to check if sslmode is specified
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		log.Fatalf("Invalid DATABASE_URL format: %v", err)
	}

	// Check if sslmode is already in the query parameters
	query := parsedURL.Query()
	if !query.Has("sslmode") {
		// Add sslmode=disable if not present
		query.Set("sslmode", "disable")
		parsedURL.RawQuery = query.Encode()
		log.Println("Note: Automatically added sslmode=disable to DATABASE_URL")
	}

	return parsedURL.String()
}

func initializeWhatsApp() {
	// Get database URL from environment
	dbURL := getDatabaseURL()

	// Create database container with PostgreSQL
	storeContainer, err := sqlstore.New(context.Background(), "postgres", dbURL, waLog.Stdout("Database", "INFO", true))
	if err != nil {
		log.Fatalf("Failed to create database container: %v", err)
	}

	// Get device store
	deviceStore, err := storeContainer.GetFirstDevice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get device store: %v", err)
	}

	// Create WhatsApp client
	clientLog := waLog.Stdout("Client", "INFO", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	// Add event handlers
	client.AddEventHandler(handler)

	// Check if already paired
	if client.Store.ID != nil {
		isPaired = true
		// Connect to existing session
		err = client.Connect()
		if err != nil {
			log.Printf("Failed to connect to existing session: %v", err)
			isPaired = false
		} else {
			log.Println("Connected to WhatsApp with existing session")
		}
	}

	// Get webhook URL from environment
	webhookURL = os.Getenv("WA_WEBHOOK_URL")
	if webhookURL != "" {
		log.Println("Webhook URL configured:", webhookURL)
	}
}

// /pair endpoint - generate QR code for pairing
func pairHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user wants image response (via Accept header or query parameter)
	acceptHeader := r.Header.Get("Accept")
	queryFormat := r.URL.Query().Get("format")

	wantsImage := (acceptHeader == "image/png") ||
		(acceptHeader != "" && acceptHeader != "application/json" && acceptHeader != "*/*") ||
		(queryFormat == "image")

	// If already paired or connected, disconnect and clear session first
	if client.IsConnected() {
		client.Disconnect()
		isPaired = false
		log.Println("Disconnected from previous session")
	}

	// Clear existing session if any
	if client.Store.ID != nil {
		err := client.Store.Delete(context.Background())
		if err != nil {
			log.Printf("Warning: Failed to clear existing session: %v", err)
		}
	}

	// Add a small delay to ensure proper disconnection
	time.Sleep(1 * time.Second)

	// Get QR channel (must be called before connecting)
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		if wantsImage {
			http.Error(w, fmt.Sprintf("Failed to get QR channel: %v", err), http.StatusInternalServerError)
		} else {
			response := APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get QR channel: %v", err),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
		return
	}

	// Connect client after getting QR channel
	err = client.Connect()
	if err != nil {
		if wantsImage {
			http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		} else {
			response := APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to connect: %v", err),
			}
			json.NewEncoder(w).Encode(response)
		}
		return
	}

	// Wait for QR code
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			qrCode := evt.Code

			if wantsImage {
				// Generate QR code as PNG image
				qr, err := qrcode.New(qrCode, qrcode.Medium)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
					return
				}

				// Set content type for PNG image
				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")

				// Encode and send the image
				err = png.Encode(w, qr.Image(256))
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to encode QR image: %v", err), http.StatusInternalServerError)
					return
				}

				log.Println("QR code image generated successfully")
				return
			} else {
				// Return JSON response with QR code info
				qrImageURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", qrCode)

				response := APIResponse{
					Success: true,
					Message: "QR code generated successfully",
					Data: map[string]interface{}{
						"qr_code":        qrCode,
						"qr_image_url":   qrImageURL,
						"image_endpoint": "/pair?format=image",
						"expires_in":     evt.Timeout,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)

				// Handle QR events in background
				go handleQREvents(qrChan)
				return
			}
		} else {
			if wantsImage {
				http.Error(w, fmt.Sprintf("QR generation error: %s", evt.Event), http.StatusInternalServerError)
			} else {
				response := APIResponse{
					Success: false,
					Message: fmt.Sprintf("QR generation error: %s", evt.Event),
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
			return
		}
	case <-time.After(10 * time.Second):
		if wantsImage {
			http.Error(w, "QR code generation timeout", http.StatusRequestTimeout)
		} else {
			response := APIResponse{
				Success: false,
				Message: "QR code generation timeout",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
		return
	}
}

func handleQREvents(qrChan <-chan whatsmeow.QRChannelItem) {
	for evt := range qrChan {
		switch evt.Event {
		case "success":
			isPaired = true
			log.Println("Successfully paired with WhatsApp!")
		case "timeout":
			log.Println("QR code pairing timed out.")
		case "err-client-outdated":
			log.Println("Client is outdated. Please update the library.")
		case "err-scanned-without-multidevice":
			log.Println("QR code was scanned but multidevice is not enabled on the phone.")
		case "error":
			log.Printf("QR pairing error: %v\n", evt.Error)
		}
	}
}

// /send endpoint - send message to a number
func sendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		response := APIResponse{
			Success: false,
			Message: "Method not allowed. Use POST",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if paired
	if !isPaired || !client.IsConnected() {
		response := APIResponse{
			Success: false,
			Message: "Not paired with WhatsApp. Please use /pair endpoint first",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req SendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response := APIResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate input
	if req.Number == "" || req.Message == "" {
		response := APIResponse{
			Success: false,
			Message: "Number and message are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse phone number (assuming it includes country code without +)
	targetJID, err := types.ParseJID(req.Number + "@s.whatsapp.net")
	if err != nil {
		response := APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid phone number: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Build message
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	// Send message
	_, err = client.SendMessage(context.Background(), targetJID, msg)
	if err != nil {
		response := APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send message: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Message sent successfully",
		Data: map[string]string{
			"number":  req.Number,
			"message": req.Message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"paired":             isPaired,
		"connected":          client != nil && client.IsConnected(),
		"webhook_configured": webhookURL != "",
	}

	response := APIResponse{
		Success: true,
		Message: "WhatsApp service is running",
		Data:    status,
	}
	json.NewEncoder(w).Encode(response)
}

// Swagger documentation endpoint
func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	swaggerInfo := map[string]interface{}{
		"title":       "WhatsApp Web API",
		"description": "REST API for WhatsApp Web integration",
		"version":     "1.0.0",
		"endpoints": map[string]string{
			"pair":    "GET  /pair   - Generate QR code for pairing",
			"send":    "POST /send   - Send message (requires pairing)",
			"health":  "GET  /health - Check service status",
			"swagger": "GET  /swagger - API documentation info",
			"docs":    "GET  /swagger.yaml - Full OpenAPI specification",
		},
		"documentation": "Full API documentation available at /swagger.yaml",
		"swagger_ui":    "Use Swagger UI with the yaml file: https://editor.swagger.io/",
	}

	response := APIResponse{
		Success: true,
		Message: "WhatsApp Web API Documentation",
		Data:    swaggerInfo,
	}
	json.NewEncoder(w).Encode(response)
}

func handler(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.Message:
		handleMessage(evt)
	case *events.Connected:
		log.Println("Connected to WhatsApp!")
	case *events.Disconnected:
		log.Println("Disconnected from WhatsApp")
		isPaired = false
	case *events.PairSuccess:
		log.Printf("Successfully paired! Device: %s\n", evt.ID)
		isPaired = true
	case *events.LoggedOut:
		log.Println("Logged out from WhatsApp")
		isPaired = false
	}
}

func handleMessage(evt *events.Message) {
	// Ignore messages from ourselves
	if evt.Info.IsFromMe {
		return
	}

	// Log message reception
	log.Printf("Received message from %s (Chat: %s)\n", evt.Info.Sender.String(), evt.Info.Chat.String())

	// Extract message content
	var messageContent string
	if evt.Message != nil {
		if evt.Message.Conversation != nil && *evt.Message.Conversation != "" {
			messageContent = *evt.Message.Conversation
		} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
			messageContent = *evt.Message.ExtendedTextMessage.Text
		} else {
			messageContent = "Non-text message received"
		}
	}

	// Send to webhook if configured
	if webhookURL != "" {
		sendToWebhook("message", messageContent, evt.Info.Sender.String(), evt.Info.Chat.String())
	}

	// Mark message as read
	err := client.MarkRead(
		[]types.MessageID{evt.Info.ID},
		time.Now(),
		evt.Info.Chat,
		evt.Info.Sender,
		types.ReceiptTypeRead,
	)
	if err != nil {
		log.Printf("Failed to mark message as read: %v", err)
	} else {
		log.Printf("Message marked as read successfully")
	}

	log.Printf("Message content: %s\n", messageContent)
}

func sendToWebhook(event, message, sender, chat string) {
	payload := WebhookPayload{
		Event:   event,
		Message: message,
		Sender:  sender,
		Chat:    chat,
		Time:    time.Now(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Webhook sent successfully to %s", webhookURL)
	} else {
		log.Printf("Webhook request failed with status: %d", resp.StatusCode)
	}
}

func main() {
	// Initialize WhatsApp client
	initializeWhatsApp()

	// Create router
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/pair", pairHandler).Methods("GET")
	r.HandleFunc("/send", sendHandler).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Serve Swagger documentation
	r.HandleFunc("/swagger", swaggerHandler).Methods("GET")
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(http.Dir("./"))))

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting WhatsApp Web API server on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /pair      - Generate QR code for pairing")
	log.Printf("  POST /send      - Send message (requires pairing)")
	log.Printf("  GET  /health    - Check service status")
	log.Printf("  GET  /swagger   - API documentation info")
	log.Printf("  GET  /swagger.yaml - Full OpenAPI specification")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Disconnect WhatsApp client
	if client != nil && client.IsConnected() {
		client.Disconnect()
		log.Println("WhatsApp client disconnected")
	}

	log.Println("Server stopped")
}
