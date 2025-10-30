package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
	"golang.org/x/image/webp"
	"google.golang.org/protobuf/proto"

	_ "github.com/lib/pq"
)

var (
	client     *whatsmeow.Client
	qrChannel  chan string
	webhookURL string
	isPaired   bool   = false
	version    string = "v1.7.0"
)

// Response structures for API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Attachment struct {
	Type     string `json:"type"`     // image, document, audio, video
	URL      string `json:"url"`      // URL or base64 data
	Filename string `json:"filename"` // optional filename for documents
	Caption  string `json:"caption"`  // optional caption
}

type SendRequest struct {
	Number      string       `json:"number"`
	Message     string       `json:"message"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type WebhookPayload struct {
	Event      string                 `json:"event"`
	Message    string                 `json:"message,omitempty"`
	Sender     string                 `json:"sender,omitempty"`
	Chat       string                 `json:"chat,omitempty"`
	Time       time.Time              `json:"time"`
	Attachment map[string]interface{} `json:"attachment,omitempty"`
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
	log.Println("=== INITIALIZING WHATSAPP CLIENT ===")

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

	// Check if already paired and attempt connection with better error handling
	if client.Store.ID != nil {
		log.Printf("Found existing session for device: %s", client.Store.ID.String())
		isPaired = true

		// Attempt to connect to existing session
		log.Println("Attempting to connect to existing session...")
		err = client.Connect()
		if err != nil {
			log.Printf("Failed to connect to existing session: %v", err)
			log.Println("💡 This could be due to:")
			log.Println("   - Device being disconnected from WhatsApp mobile app")
			log.Println("   - Device limit exceeded on WhatsApp account")
			log.Println("   - Network connectivity issues")
			log.Println("   - Session corruption")
			log.Println("💡 Use /pair endpoint to create a new session")
			isPaired = false
		} else {
			log.Println("🟢 Successfully connected to WhatsApp with existing session")
		}
	} else {
		log.Println("No existing session found - use /pair endpoint to create one")
	}

	// Get webhook URL from environment
	webhookURL = os.Getenv("WA_WEBHOOK_URL")
	if webhookURL != "" {
		log.Println("Webhook URL configured:", webhookURL)
	}

	log.Println("=== WHATSAPP CLIENT INITIALIZATION COMPLETE ===")
}

// /pair endpoint - generate QR code for pairing
func pairHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=== PAIRING REQUEST STARTED ===")

	// If already paired or connected, disconnect and clear session first
	if client != nil && client.IsConnected() {
		log.Println("Disconnecting existing session...")
		client.Disconnect()
		isPaired = false
		log.Println("Disconnected from previous session")
	}

	// Clear existing session if any
	if client != nil && client.Store != nil && client.Store.ID != nil {
		log.Printf("Clearing existing session for device: %s", client.Store.ID.String())
		err := client.Store.Delete(context.Background())
		if err != nil {
			log.Printf("Warning: Failed to clear existing session: %v", err)
		} else {
			log.Println("Existing session cleared successfully")
		}
	}

	// Add a small delay to ensure proper disconnection
	time.Sleep(2 * time.Second)

	// Get QR channel (must be called before connecting)
	log.Println("Getting QR channel...")
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		log.Printf("Failed to get QR channel: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get QR channel: %v", err), http.StatusInternalServerError)
		return
	}

	// Connect client after getting QR channel
	log.Println("Connecting to WhatsApp...")
	err = client.Connect()
	if err != nil {
		log.Printf("Failed to connect: %v", err)
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("Connected successfully, waiting for QR code...")

	// Wait for QR code with extended timeout
	select {
	case evt := <-qrChan:
		log.Printf("QR event received: %s", evt.Event)
		if evt.Event == "code" {
			qrCode := evt.Code
			log.Printf("QR code generated, length: %d", len(qrCode))

			// Generate QR code as PNG image
			qr, err := qrcode.New(qrCode, qrcode.Medium)
			if err != nil {
				log.Printf("Failed to generate QR code: %v", err)
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
				log.Printf("Failed to encode QR image: %v", err)
				http.Error(w, fmt.Sprintf("Failed to encode QR image: %v", err), http.StatusInternalServerError)
				return
			}

			log.Println("QR code image generated successfully")
			log.Println("=== PAIRING REQUEST COMPLETED ===")

			// Handle QR events in background
			go handleQREvents(qrChan)
			return
		} else {
			log.Printf("QR generation error: %s", evt.Event)
			if evt.Error != nil {
				log.Printf("QR error details: %v", evt.Error)
			}
			http.Error(w, fmt.Sprintf("QR generation error: %s", evt.Event), http.StatusInternalServerError)
			return
		}
	case <-time.After(15 * time.Second):
		log.Println("QR code generation timeout after 15 seconds")
		http.Error(w, "QR code generation timeout - please try again", http.StatusRequestTimeout)
		return
	}
}

func handleQREvents(qrChan <-chan whatsmeow.QRChannelItem) {
	log.Println("=== QR EVENT HANDLER STARTED ===")
	for evt := range qrChan {
		log.Printf("QR Event: %s", evt.Event)
		switch evt.Event {
		case "success":
			isPaired = true
			log.Println("🎉 Successfully paired with WhatsApp!")
			log.Printf("Device ID: %s", client.Store.ID.String())
		case "timeout":
			log.Println("⏰ QR code pairing timed out.")
			log.Println("💡 Tips: Check if WhatsApp is open on your phone and try scanning again")
		case "err-client-outdated":
			log.Println("🔄 Client is outdated. Please update the application.")
		case "err-scanned-without-multidevice":
			log.Println("📱 QR code was scanned but multi-device is not enabled on the phone.")
			log.Println("💡 Solution: Go to WhatsApp Settings > Linked Devices > Enable multi-device")
		case "err-device-limit-exceeded":
			log.Println("🚫 Device limit exceeded on WhatsApp account.")
			log.Println("💡 Solution: Remove unused devices from WhatsApp Settings > Linked Devices")
		case "err-already-connected":
			log.Println("⚠️ Device is already connected to another session.")
			log.Println("💡 Solution: Disconnect other devices first")
		case "error":
			log.Printf("❌ QR pairing error: %v", evt.Error)
			if evt.Error != nil {
				log.Printf("Error details: %s", evt.Error.Error())
			}
		default:
			log.Printf("❓ Unknown QR event: %s", evt.Event)
		}
	}
	log.Println("=== QR EVENT HANDLER ENDED ===")
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
	if req.Number == "" {
		response := APIResponse{
			Success: false,
			Message: "Number is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Message == "" && len(req.Attachments) == 0 {
		response := APIResponse{
			Success: false,
			Message: "Either message or attachments are required",
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

	var messages []*waProto.Message

	// Check if we have text + single image attachment to combine
	if req.Message != "" && len(req.Attachments) == 1 && req.Attachments[0].Type == "image" {
		// Combine text as image caption
		attachment := req.Attachments[0]
		attachment.Caption = req.Message // Use text message as caption
		attachmentMsg, err := prepareAttachmentMessage(attachment, targetJID)
		if err != nil {
			response := APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to prepare attachment: %v", err),
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		messages = append(messages, attachmentMsg)
	} else {
		// Add text message if provided
		if req.Message != "" {
			messages = append(messages, &waProto.Message{
				Conversation: proto.String(req.Message),
			})
		}

		// Process attachments
		for _, attachment := range req.Attachments {
			attachmentMsg, err := prepareAttachmentMessage(attachment, targetJID)
			if err != nil {
				response := APIResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to prepare attachment: %v", err),
				}
				json.NewEncoder(w).Encode(response)
				return
			}
			messages = append(messages, attachmentMsg)
		}
	}

	// Send typing indicator before sending messages
	sendTypingIndicator(targetJID)

	// Send all messages
	var sentMessages []map[string]interface{}
	for i, msg := range messages {
		_, err = client.SendMessage(context.Background(), targetJID, msg)
		if err != nil {
			response := APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message %d: %v", i+1, err),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		sentInfo := map[string]interface{}{"index": i + 1}
		if req.Message != "" && len(req.Attachments) == 1 && req.Attachments[0].Type == "image" {
			// Combined message case
			sentInfo["type"] = "image_with_caption"
			sentInfo["content"] = req.Message
			sentInfo["filename"] = req.Attachments[0].Filename
		} else if i == 0 && req.Message != "" {
			sentInfo["type"] = "text"
			sentInfo["content"] = req.Message
		} else if i > 0 || req.Message == "" {
			attachmentIndex := i
			if req.Message != "" {
				attachmentIndex = i - 1
			}
			if attachmentIndex < len(req.Attachments) {
				sentInfo["type"] = req.Attachments[attachmentIndex].Type
				sentInfo["filename"] = req.Attachments[attachmentIndex].Filename
			}
		}
		sentMessages = append(sentMessages, sentInfo)
	}

	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully sent %d message(s)", len(messages)),
		Data: map[string]interface{}{
			"number":      req.Number,
			"message":     req.Message,
			"attachments": req.Attachments,
			"sent":        sentMessages,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"version":            version,
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

// Device management endpoint
func devicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if client == nil {
		response := APIResponse{
			Success: false,
			Message: "Client not initialized",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	deviceInfo := map[string]interface{}{
		"connected": client.IsConnected(),
		"paired":    isPaired,
	}

	// Only add device info if store exists and has ID
	if client.Store != nil && client.Store.ID != nil {
		deviceInfo["device_id"] = client.Store.ID.String()
		deviceInfo["jid"] = client.Store.ID
		deviceInfo["phone"] = client.Store.ID.User
	} else {
		deviceInfo["device_id"] = nil
		deviceInfo["jid"] = nil
		deviceInfo["phone"] = nil
	}

	response := APIResponse{
		Success: true,
		Message: "Device information retrieved",
		Data:    deviceInfo,
	}
	json.NewEncoder(w).Encode(response)
}

// Disconnect endpoint
func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if client == nil {
		response := APIResponse{
			Success: false,
			Message: "Client not initialized",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Disconnect if connected
	if client.IsConnected() {
		client.Disconnect()
		log.Println("Manually disconnected from WhatsApp")
	}

	// Clear session
	if client.Store.ID != nil {
		err := client.Store.Delete(context.Background())
		if err != nil {
			log.Printf("Warning: Failed to clear session during disconnect: %v", err)
		} else {
			log.Println("Session cleared successfully")
		}
	}

	isPaired = false

	response := APIResponse{
		Success: true,
		Message: "Successfully disconnected and session cleared",
	}
	json.NewEncoder(w).Encode(response)
}

// Image endpoint - serve downloaded images
func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Security check: ensure filename doesn't contain path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Construct file path
	filePath := fmt.Sprintf("downloads/%s", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Determine content type
	ext := strings.ToLower(filepath.Ext(filename))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	default:
		contentType = "application/octet-stream"
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Serve file
	http.ServeFile(w, r, filePath)
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
			"send":    "POST /send   - Send message with attachments (requires pairing)",
			"health":  "GET  /health - Check service status",
			"images":  "GET  /images/{filename} - Serve downloaded images",
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
		log.Println("🟢 Connected to WhatsApp!")
		if client.Store.ID != nil {
			log.Printf("Device ID: %s", client.Store.ID.String())
		}
	case *events.Disconnected:
		log.Println("🔴 Disconnected from WhatsApp")
		isPaired = false
	case *events.PairSuccess:
		log.Printf("🎉 Successfully paired! Device: %s", evt.ID)
		isPaired = true
	case *events.LoggedOut:
		log.Println("🔒 Logged out from WhatsApp")
		log.Println("💡 This may happen if another device connects or if you log out from WhatsApp mobile app")
		isPaired = false
	case *events.StreamError:
		log.Printf("🚫 Stream error occurred")
		log.Println("💡 This may indicate connection issues or device limit problems")
	case *events.ConnectFailure:
		log.Printf("❌ Connection failed: %v", evt.Reason)
		log.Println("💡 Check your internet connection and WhatsApp device limits")
	}
}

func logMessageDetails(evt *events.Message) {
	// Log basic message information
	log.Printf("=== MESSAGE RECEIVED ===")
	log.Printf("Message ID: %s", evt.Info.ID)
	log.Printf("From: %s", evt.Info.Sender.String())
	log.Printf("Chat: %s", evt.Info.Chat.String())
	log.Printf("Timestamp: %s", evt.Info.Timestamp)
	log.Printf("Push Name: %s", evt.Info.PushName)
	log.Printf("Is From Me: %t", evt.Info.IsFromMe)
	log.Printf("Is Group: %t", evt.Info.Chat.Server == "g.us")

	// Log message type information
	if evt.Message != nil {
		log.Printf("Message Type Analysis:")
		if evt.Message.Conversation != nil && *evt.Message.Conversation != "" {
			log.Printf("  - Text message: %s", *evt.Message.Conversation)
		}
		if evt.Message.ExtendedTextMessage != nil {
			log.Printf("  - Extended text message")
		}
		if evt.Message.ImageMessage != nil {
			log.Printf("  - Image message")
		}
		if evt.Message.DocumentMessage != nil {
			log.Printf("  - Document message")
		}
		if evt.Message.AudioMessage != nil {
			log.Printf("  - Audio message")
		}
		if evt.Message.VideoMessage != nil {
			log.Printf("  - Video message")
		}
		if evt.Message.StickerMessage != nil {
			log.Printf("  - Sticker message")
		}
		if evt.Message.ContactMessage != nil {
			log.Printf("  - Contact message")
		}
		if evt.Message.LocationMessage != nil {
			log.Printf("  - Location message")
		}
		if evt.Message.ReactionMessage != nil {
			log.Printf("  - Reaction message")
		}
		if evt.Message.ButtonsResponseMessage != nil {
			log.Printf("  - Button response message")
		}
		if evt.Message.ListResponseMessage != nil {
			log.Printf("  - List response message")
		}
		if evt.Message.PollCreationMessage != nil {
			log.Printf("  - Poll creation message")
		}
		if evt.Message.PollUpdateMessage != nil {
			log.Printf("  - Poll update message")
		}
	} else {
		log.Printf("  - Message content is nil")
	}

	log.Printf("========================")
}

func handleMessage(evt *events.Message) {
	// Ignore messages from ourselves
	if evt.Info.IsFromMe {
		return
	}

	// Log comprehensive message information
	logMessageDetails(evt)

	// Mark message as read FIRST
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

	// Extract message content and handle automatic image download
	var messageContent string
	var attachmentInfo map[string]interface{}

	if evt.Message != nil {
		if evt.Message.Conversation != nil && *evt.Message.Conversation != "" {
			messageContent = *evt.Message.Conversation
		} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
			messageContent = *evt.Message.ExtendedTextMessage.Text
		} else if evt.Message.ImageMessage != nil {
			// Handle image message
			imgMsg := evt.Message.ImageMessage
			caption := ""
			if imgMsg.Caption != nil {
				caption = *imgMsg.Caption
			}
			messageContent = fmt.Sprintf("Image received%s", func() string {
				if caption != "" {
					return fmt.Sprintf(": %s", caption)
				}
				return ""
			}())

			// Automatically download the image
			go func() {
				err := downloadAndSaveImage(evt.Info.ID, imgMsg)
				if err != nil {
					log.Printf("Failed to download image: %v", err)
				} else {
					log.Printf("Image downloaded successfully")
				}
			}()

			// Store image info for webhook and logging
			attachmentInfo = map[string]interface{}{
				"type":        "image",
				"caption":     caption,
				"mimetype":    imgMsg.Mimetype,
				"file_length": imgMsg.FileLength,
				"width":       imgMsg.Width,
				"height":      imgMsg.Height,
				"url":         fmt.Sprintf("/images/%s.jpg", evt.Info.ID),
			}
		} else if evt.Message.DocumentMessage != nil {
			docMsg := evt.Message.DocumentMessage
			title := ""
			if docMsg.Title != nil {
				title = *docMsg.Title
			}
			messageContent = fmt.Sprintf("Document received: %s", title)
			attachmentInfo = map[string]interface{}{
				"type":        "document",
				"title":       title,
				"mimetype":    docMsg.Mimetype,
				"file_length": docMsg.FileLength,
				"page_count":  docMsg.PageCount,
			}
		} else if evt.Message.AudioMessage != nil {
			audioMsg := evt.Message.AudioMessage
			messageContent = "Audio message received"
			attachmentInfo = map[string]interface{}{
				"type":        "audio",
				"mimetype":    audioMsg.Mimetype,
				"file_length": audioMsg.FileLength,
				"seconds":     audioMsg.Seconds,
			}
		} else if evt.Message.VideoMessage != nil {
			vidMsg := evt.Message.VideoMessage
			caption := ""
			if vidMsg.Caption != nil {
				caption = *vidMsg.Caption
			}
			messageContent = fmt.Sprintf("Video received%s", func() string {
				if caption != "" {
					return fmt.Sprintf(": %s", caption)
				}
				return ""
			}())
			attachmentInfo = map[string]interface{}{
				"type":        "video",
				"caption":     caption,
				"mimetype":    vidMsg.Mimetype,
				"file_length": vidMsg.FileLength,
				"seconds":     vidMsg.Seconds,
				"width":       vidMsg.Width,
				"height":      vidMsg.Height,
			}
		} else if evt.Message.StickerMessage != nil {
			stickerMsg := evt.Message.StickerMessage
			messageContent = "Sticker received"
			attachmentInfo = map[string]interface{}{
				"type":        "sticker",
				"mimetype":    stickerMsg.Mimetype,
				"file_length": stickerMsg.FileLength,
				"width":       stickerMsg.Width,
				"height":      stickerMsg.Height,
			}
		} else if evt.Message.ContactMessage != nil {
			contactMsg := evt.Message.ContactMessage
			messageContent = fmt.Sprintf("Contact received: %s", contactMsg.DisplayName)
			attachmentInfo = map[string]interface{}{
				"type":         "contact",
				"display_name": contactMsg.DisplayName,
				"vcard":        contactMsg.Vcard,
			}
		} else if evt.Message.LocationMessage != nil {
			locMsg := evt.Message.LocationMessage
			messageContent = fmt.Sprintf("Location received: %s", locMsg.Name)
			attachmentInfo = map[string]interface{}{
				"type":      "location",
				"name":      locMsg.Name,
				"address":   locMsg.Address,
				"latitude":  locMsg.DegreesLatitude,
				"longitude": locMsg.DegreesLongitude,
			}
		} else {
			messageContent = "Non-text message received"
			attachmentInfo = map[string]interface{}{
				"type": "unknown",
			}
		}
	}

	// Log the processed message content and attachment details
	log.Printf("Processed message - Content: %s", messageContent)
	if attachmentInfo != nil {
		log.Printf("Attachment details: %+v", attachmentInfo)
	}

	// Send to webhook if configured
	if webhookURL != "" {
		sendToWebhook("message", messageContent, evt.Info.Sender.String(), evt.Info.Chat.String(), attachmentInfo)
	}
}

func downloadAndSaveImage(messageID types.MessageID, imgMsg *waProto.ImageMessage) error {
	log.Printf("=== IMAGE DOWNLOAD START ===")
	log.Printf("Message ID: %s", messageID)
	log.Printf("Image URL: %s", *imgMsg.URL)
	log.Printf("Direct Path: %s", *imgMsg.DirectPath)
	log.Printf("Mimetype: %s", *imgMsg.Mimetype)
	log.Printf("File Length: %d bytes", *imgMsg.FileLength)
	if imgMsg.Width != nil && imgMsg.Height != nil {
		log.Printf("Dimensions: %dx%d", *imgMsg.Width, *imgMsg.Height)
	}
	if imgMsg.Caption != nil && *imgMsg.Caption != "" {
		log.Printf("Caption: %s", *imgMsg.Caption)
	}

	if imgMsg.URL == nil || imgMsg.DirectPath == nil {
		return fmt.Errorf("image URL or DirectPath is nil")
	}

	// Download the image using the client's built-in downloader
	log.Printf("Starting download...")
	data, err := client.Download(context.Background(), imgMsg)
	if err != nil {
		log.Printf("Download failed: %v", err)
		return fmt.Errorf("failed to download image: %v", err)
	}

	log.Printf("Successfully downloaded image data: %d bytes", len(data))

	// Optionally save to file (you can customize this path)
	filename := fmt.Sprintf("downloads/%s.jpg", messageID)
	log.Printf("Creating downloads directory if needed...")
	err = os.MkdirAll("downloads", 0755)
	if err != nil {
		log.Printf("Failed to create downloads directory: %v", err)
		return fmt.Errorf("failed to create downloads directory: %v", err)
	}

	log.Printf("Saving image to: %s", filename)
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Printf("Failed to save image file: %v", err)
		return fmt.Errorf("failed to save image file: %v", err)
	}

	log.Printf("Image successfully saved to: %s", filename)
	log.Printf("=== IMAGE DOWNLOAD COMPLETE ===")
	return nil
}

func downloadFile(url string) ([]byte, string, error) {
	log.Printf("=== FILE DOWNLOAD START ===")
	log.Printf("Downloading from URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP GET request failed: %v", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	log.Printf("HTTP Response Status: %d", resp.StatusCode)
	log.Printf("Content-Type: %s", resp.Header.Get("Content-Type"))
	log.Printf("Content-Length: %s", resp.Header.Get("Content-Length"))

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP error: %d", resp.StatusCode)
		return nil, "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	log.Printf("Reading response body...")
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, "", err
	}

	log.Printf("Successfully downloaded %d bytes", len(data))

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
		log.Printf("Detected content type: %s", contentType)
	} else {
		log.Printf("Server content type: %s", contentType)
	}

	log.Printf("=== FILE DOWNLOAD COMPLETE ===")
	return data, contentType, nil
}

func convertImageToJPEG(data []byte, contentType string) ([]byte, error) {
	// If already JPEG, return as-is
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		return data, nil
	}

	// Decode image
	var img image.Image
	var err error

	switch {
	case strings.Contains(contentType, "png"):
		img, err = png.Decode(bytes.NewReader(data))
	case strings.Contains(contentType, "webp"):
		img, err = webp.Decode(bytes.NewReader(data))
	default:
		// Try to decode as generic image
		img, _, err = image.Decode(bytes.NewReader(data))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Check if image is valid before encoding
	if img == nil {
		return nil, fmt.Errorf("decoded image is nil")
	}

	// Encode as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode as JPEG: %v", err)
	}

	log.Printf("Successfully converted %s to JPEG", contentType)
	return buf.Bytes(), nil
}

func sendTypingIndicator(targetJID types.JID) {
	// Send chat state (composing) to indicate typing
	chatJID := targetJID.ToNonAD()
	if chatJID.Server == "g.us" {
		// For group chats, use the group JID directly
		chatJID = targetJID
	}

	// Send composing presence
	err := client.SendChatPresence(chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	if err != nil {
		log.Printf("Failed to send typing indicator: %v", err)
	} else {
		log.Printf("Typing indicator sent to %s", chatJID.String())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func prepareAttachmentMessage(attachment Attachment, targetJID types.JID) (*waProto.Message, error) {
	log.Printf("=== ATTACHMENT PREPARATION ===")
	log.Printf("Attachment Type: %s", attachment.Type)
	log.Printf("Attachment URL: %s", attachment.URL)
	log.Printf("Attachment Caption: %s", attachment.Caption)
	log.Printf("Attachment Filename: %s", attachment.Filename)
	log.Printf("Target JID: %s", targetJID.String())

	var data []byte
	var contentType string
	var err error

	if strings.HasPrefix(attachment.URL, "http") {
		data, contentType, err = downloadFile(attachment.URL)
	} else {
		return nil, fmt.Errorf("attachment URL must be a publicly accessible HTTP/HTTPS link, not base64 data. Found: %s", attachment.URL[:min(50, len(attachment.URL))])
	}

	if err != nil {
		log.Printf("Failed to load attachment: %v", err)
		return nil, fmt.Errorf("failed to load attachment: %v", err)
	}

	log.Printf("Attachment loaded successfully: %d bytes, content type: %s", len(data), contentType)

	// Convert image to JPEG if needed
	if attachment.Type == "image" {
		log.Printf("Converting image to JPEG...")
		data, err = convertImageToJPEG(data, contentType)
		if err != nil {
			log.Printf("Failed to convert image: %v", err)
			return nil, fmt.Errorf("failed to convert image: %v", err)
		}
		contentType = "image/jpeg"
		log.Printf("Image converted to JPEG successfully")
	}

	var mediaType whatsmeow.MediaType
	switch attachment.Type {
	case "image":
		mediaType = whatsmeow.MediaImage
	case "document":
		mediaType = whatsmeow.MediaDocument
	case "audio":
		mediaType = whatsmeow.MediaAudio
	case "video":
		mediaType = whatsmeow.MediaVideo
	default:
		mediaType = whatsmeow.MediaDocument
	}

	log.Printf("Uploading attachment to WhatsApp servers...")
	uploaded, err := client.Upload(context.Background(), data, mediaType)
	if err != nil {
		log.Printf("Failed to upload attachment: %v", err)
		return nil, fmt.Errorf("failed to upload attachment: %v", err)
	}

	log.Printf("Attachment uploaded successfully")
	log.Printf("Upload URL: %s", uploaded.URL)
	log.Printf("Direct Path: %s", uploaded.DirectPath)

	var message *waProto.Message
	switch attachment.Type {
	case "image":
		message = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				Mimetype:      proto.String(contentType),
				Caption:       proto.String(attachment.Caption),
				FileLength:    proto.Uint64(uint64(len(data))),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
			},
		}
		log.Printf("Image message prepared successfully")
	case "document":
		filename := attachment.Filename
		if filename == "" {
			filename = "document"
		}
		message = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				Mimetype:      proto.String(contentType),
				Title:         proto.String(filename),
				FileName:      proto.String(filename),
				FileLength:    proto.Uint64(uint64(len(data))),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
			},
		}
		log.Printf("Document message prepared successfully")
	case "audio":
		message = &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				Mimetype:      proto.String(contentType),
				FileLength:    proto.Uint64(uint64(len(data))),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
			},
		}
		log.Printf("Audio message prepared successfully")
	case "video":
		message = &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				Mimetype:      proto.String(contentType),
				Caption:       proto.String(attachment.Caption),
				FileLength:    proto.Uint64(uint64(len(data))),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
			},
		}
		log.Printf("Video message prepared successfully")
	default:
		log.Printf("Unsupported attachment type: %s", attachment.Type)
		return nil, fmt.Errorf("unsupported attachment type: %s", attachment.Type)
	}

	log.Printf("=== ATTACHMENT PREPARATION COMPLETE ===")
	return message, nil
}

func sendToWebhook(event, message, sender, chat string, attachment map[string]interface{}) {
	log.Printf("=== WEBHOOK SENDING ===")
	log.Printf("Event: %s", event)
	log.Printf("Sender: %s", sender)
	log.Printf("Chat: %s", chat)
	log.Printf("Message: %s", message)
	log.Printf("Webhook URL: %s", webhookURL)

	if attachment != nil {
		log.Printf("Attachment: %+v", attachment)
	}

	payload := WebhookPayload{
		Event:      event,
		Message:    message,
		Sender:     sender,
		Chat:       chat,
		Time:       time.Now(),
		Attachment: attachment,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}

	log.Printf("Webhook payload size: %d bytes", len(jsonData))
	log.Printf("Sending webhook request...")

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook response status: %d", resp.StatusCode)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Webhook sent successfully to %s", webhookURL)
	} else {
		log.Printf("Webhook request failed with status: %d", resp.StatusCode)
	}
	log.Printf("=== WEBHOOK COMPLETE ===")
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
	r.HandleFunc("/devices", devicesHandler).Methods("GET")
	r.HandleFunc("/disconnect", disconnectHandler).Methods("POST")
	r.HandleFunc("/images/{filename}", imageHandler).Methods("GET")

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
	log.Printf("  POST /send      - Send message with attachments (requires pairing)")
	log.Printf("  GET  /health    - Check service status")
	log.Printf("  GET  /devices   - Get device information")
	log.Printf("  POST /disconnect - Disconnect and clear session")
	log.Printf("  GET  /images/{filename} - Serve downloaded images")
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
