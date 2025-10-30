# WhatsApp Web API

🚀 A production-ready REST API for WhatsApp Web integration that enables programmatic sending of messages, QR code pairing, and webhook-based message reception.

## 📋 Changelog

### **v1.7.0** (Latest) - Enhanced Troubleshooting & Testing Tools
- 📖 **Comprehensive Troubleshooting Guide**: New `TROUBLESHOOTING.md` with detailed solutions for common issues
- 🧪 **Endpoint Testing Script**: New `test_endpoints.sh` for automated API endpoint testing
- 🔧 **Enhanced Device Management**: Improved device information and session control endpoints
- 📊 **Rich Logging System**: Comprehensive emoji-based logging for better debugging and monitoring
- 🚨 **Advanced Error Handling**: Better detection and guidance for WhatsApp device limits and connection issues
- ⏰ **Optimized Timeouts**: Extended QR code timeout and improved connection handling
- 🛠️ **Session Management**: Enhanced session cleanup and recovery mechanisms
- 💡 **Built-in Guidance**: Actionable suggestions and tips directly in error messages and logs
- 🎯 **Production Ready**: Improved stability and reliability for production deployments

### **v1.6.1** - Device Management & Enhanced Error Handling
- 🔧 **Device Management**: New `GET /devices` endpoint to view current device information
- 🔄 **Session Control**: New `POST /disconnect` endpoint to manually disconnect and clear sessions
- 📊 **Enhanced Logging**: Comprehensive logging with emojis for better debugging
- 🚫 **Device Limit Detection**: Better error messages for WhatsApp device limit issues
- ⏰ **Extended QR Timeout**: Increased QR code generation timeout to 15 seconds
- 🛠️ **Improved Session Cleanup**: Better handling of stale sessions on startup
- 💡 **Troubleshooting Tips**: Built-in guidance for common pairing issues
- 🎯 **Error Recovery**: Enhanced error handling with actionable suggestions

### **v1.6.0** - Image Endpoint & Enhanced Webhooks
- 🖼️ **Image Endpoint**: New `GET /images/{filename}` endpoint to serve downloaded images
- 📡 **Enhanced Webhooks**: Image URLs now included in webhook attachment payloads
- 📊 **Rich Metadata**: Comprehensive attachment metadata for all message types (images, documents, audio, video, stickers, contacts, locations)
- 🔒 **Security Improvements**: Path traversal protection and input validation for image endpoint
- 📝 **Better Logging**: Enhanced logging throughout the application with detailed message information
- 🎯 **Content Type Detection**: Automatic MIME type detection for served images
- ⚡ **Performance**: Optimized caching headers for image serving

### **v1.5.0** - Enhanced Message Handling
- 📊 **Detailed Logging**: Comprehensive message type analysis and logging
- 🔄 **Improved Downloads**: Enhanced image download with detailed progress tracking
- 📝 **Better Error Handling**: Improved error messages and debugging information
- 🎯 **Attachment Processing**: Enhanced attachment preparation with better logging

### **v1.4.0** - WebP Support & Typing Indicators
- 🖼️ **WebP Support**: Added WebP image format support for incoming images
- ⌨️ **Typing Indicators**: Send typing indicators before sending messages
- 🔄 **Enhanced Conversion**: Improved image conversion and processing
- 📊 **Better Metadata**: Enhanced attachment metadata handling

### **v1.3.0** - Enhanced Image Sending
- ✨ **Combined Messages**: Text + single image now sends as one message with caption
- 🐛 **Fixed Image Display**: Resolved image rendering issues with proper DirectPath handling
- 🔗 **URL-Only Attachments**: Removed base64 support, requires HTTP/HTTPS URLs only
- 📦 **Linux Builds**: Automated releases for amd64, arm64, 386 architectures
- 🔧 **Simplified QR Code**: Endpoint now returns PNG image directly
- 📊 **Version Info**: Added version display to health endpoint

### **v1.2.0** - Attachment Support
- ✨ Added support for images, documents, audio, and video attachments
- 🪝 Enhanced webhook functionality
- 📚 Comprehensive API documentation

### **v1.1.0** - Core Features
- ✨ Basic message sending and QR code pairing
- 🗄️ PostgreSQL session storage
- 🔄 Session management

## ✨ Features

- **📱 QR Code Pairing**: Generate QR codes to pair WhatsApp accounts
- **💬 Message Sending**: Send text messages and attachments (images, documents, audio, video) to any WhatsApp number via REST API
- **🖼️ Smart Image Handling**: Combines text + single image into one message with caption
- **🖼️ Image Serving**: New endpoint to serve downloaded images with URL access
- **🪝 Enhanced Webhooks**: Rich webhook payloads with attachment metadata and image URLs
- **📊 Rich Metadata**: Comprehensive attachment information for all message types
- **🗄️ PostgreSQL Storage**: Secure WhatsApp session persistence in PostgreSQL
- **🔒 Auto SSL Handling**: Automatically configures PostgreSQL SSL mode
- **📚 Complete Documentation**: Full OpenAPI 3.0 / Swagger specification
- **🔄 Session Management**: Automatic reconnection and session handling
- **⚡ High Performance**: Concurrent message handling and graceful shutdown
- **🔗 URL-Only Attachments**: Secure attachment handling via HTTP/HTTPS URLs only
- **🛡️ Security**: Path traversal protection and input validation
- **📝 Detailed Logging**: Comprehensive logging for debugging and monitoring

## 📋 Requirements

### **For Binary Users** (Recommended):
- PostgreSQL 12+ (or cloud PostgreSQL service)
- WhatsApp account with multi-device enabled
- Linux/macOS/Windows operating system

### **For Developers** (Build from source):
- Go 1.24+
- PostgreSQL 12+
- WhatsApp account with multi-device enabled

## ⚙️ Environment Variables

Create a `.env` file in the project root:

```env
# Required: PostgreSQL database connection
DATABASE_URL=postgres://username:password@localhost:5432/whatsapp_db

# Optional: Server port (defaults to 8080)
PORT=8080

# Optional: Webhook URL to receive incoming messages
WA_WEBHOOK_URL=https://your-webhook-endpoint.com/webhook
```

### Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE whatsapp_db;
CREATE USER whatsapp_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE whatsapp_db TO whatsapp_user;
```

## 🚀 Quick Start

### **Option 1: Download Pre-built Binary (Recommended) 📦**

No Go installation required! Just download the binary for your platform:

1. **Download the binary**:
   ```bash
   # Linux (AMD64 - most common servers)
   wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.0.0/whatsapp-web-api-linux-amd64.zip
   unzip whatsapp-web-api-linux-amd64.zip
   chmod +x whatsapp-web-api-linux-amd64

   # macOS (Intel)
   curl -L -o whatsapp-web-api-darwin-amd64.zip https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.0.0/whatsapp-web-api-darwin-amd64.zip
   unzip whatsapp-web-api-darwin-amd64.zip
   chmod +x whatsapp-web-api-darwin-amd64

   # Windows (64-bit)
   # Download: https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.0.0/whatsapp-web-api-windows-amd64.exe.zip
   # Extract and run whatsapp-web-api-windows-amd64.exe
   ```

2. **Configure environment**:
   ```bash
   # Copy example configuration
   cp .env.example .env

   # Edit .env with your database details
   nano .env
   ```

3. **Run the server**:
   ```bash
   # Linux/macOS
   ./whatsapp-web-api-linux-amd64
   # or
   ./whatsapp-web-api-darwin-amd64

   # Windows
   whatsapp-web-api-windows-amd64.exe
   ```

4. **Verify it's working**:
   ```bash
   curl http://localhost:8080/health
   ```

### **Option 2: Build from Source** 👨‍💻

For developers who want to modify or build from source:

1. **Clone and setup**:
   ```bash
   git clone https://github.com/rizrmd/whatsapp-web-api.git
   cd whatsapp-web-api
   cp .env.example .env
   # Edit .env with your credentials
   ```

2. **Install Go** (if not installed):
   ```bash
   # Ubuntu/Debian
   sudo apt install golang-go

   # macOS
   brew install go

   # Or download from: https://golang.org/dl/
   ```

3. **Install dependencies and build**:
   ```bash
   go mod tidy
   go build -o whatsapp-web-api .
   ./whatsapp-web-api
   ```

## 📖 API Endpoints

### 1. Health Check
```http
GET /health
```

Check if the WhatsApp service is running and get current status.

**Response**:
```json
{
  "success": true,
  "message": "WhatsApp service is running",
  "data": {
    "paired": true,
    "connected": true,
    "webhook_configured": true
  }
}
```

### 2. Pair WhatsApp Device
```http
GET /pair
```

Generate a QR code to pair a new WhatsApp device. This will disconnect any existing session.

**Response**:
```json
{
  "success": true,
  "message": "QR code generated successfully",
  "data": {
    "qr_code": "3-4|1|2|3|4|5|6|7|8|9|0|1|2|3|4|5|6|7|8|9|0",
    "qr_image_url": "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=...",
    "expires_in": 60
  }
}
```

**Usage**:
1. Call `/pair` endpoint
2. Scan QR code with WhatsApp (Settings > Linked Devices > Link a device)
3. Wait for successful pairing

### 3. Send Message
```http
POST /send
Content-Type: application/json
```

Send a text message and/or attachments to a WhatsApp number.

**Smart Message Handling**:
- **Text + Single Image** → Combined into one message (text becomes image caption)
- **Text + Multiple Images** → Sent as separate messages
- **Only Text** → Text message
- **Only Image** → Image message

**Request Body**:
```json
{
  "number": "1234567890",
  "message": "Check out this amazing photo!",
  "attachments": [
    {
      "type": "image",
      "url": "https://picsum.photos/800/600",
      "caption": "This caption will be replaced by the message text"
    }
  ]
}
```

**Parameters**:
- `number` (string, required): Phone number with country code (no '+' prefix)
- `message` (string, optional): Message text (max 4096 characters)
- `attachments` (array, optional): Array of attachment objects
  - `type` (string, required): Attachment type - "image", "document", "audio", "video"
  - `url` (string, required): **Publicly accessible HTTP/HTTPS URL** for the attachment
  - `filename` (string, optional): Filename for documents
  - `caption` (string, optional): Caption for images/videos (ignored for single image + text)

**Response**:
```json
{
  "success": true,
  "message": "Successfully sent 3 message(s)",
  "data": {
    "number": "1234567890",
    "message": "Hello from WhatsApp API!",
    "attachments": [...],
    "sent": [
      {"index": 1, "type": "text", "content": "Hello from WhatsApp API!"},
      {"index": 2, "type": "image", "filename": ""},
      {"index": 3, "type": "document", "filename": "document.pdf"}
    ]
  }
}
```

### 4. Device Information
```http
GET /devices
```

Get information about the currently connected WhatsApp device.

**Response**:
```json
{
  "success": true,
  "message": "Device information retrieved",
  "data": {
    "device_id": "1234567890@s.whatsapp.net",
    "jid": "1234567890@s.whatsapp.net",
    "connected": true,
    "paired": true
  }
}
```

### 5. Disconnect Session
```http
POST /disconnect
```

Manually disconnect from WhatsApp and clear the current session. Useful for troubleshooting or when you need to pair a new device.

**Response**:
```json
{
  "success": true,
  "message": "Successfully disconnected and session cleared"
}
```

### 6. Image Serving
```http
GET /images/{filename}
```

Access downloaded images via secure endpoint. Images are automatically downloaded when received and can be accessed through this endpoint.

**Parameters**:
- `filename` (string, required): Image filename (e.g., `ABC123.jpg`)

**Response**: Image file with appropriate content type headers

**Example**: `http://localhost:8080/images/ABC123.jpg`

### 7. API Documentation
```http
GET /swagger
GET /swagger.yaml
```

Access API documentation and OpenAPI specification.

## 💻 Binary Usage Guide

### **Step-by-Step Binary Setup**

1. **Download the right binary for your system**:

| Platform | Architecture | Download Command |
|----------|-------------|------------------|
| **Linux** | AMD64 (most servers) | `wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-linux-amd64.zip` |
| **Linux** | ARM64 (AWS Graviton) | `wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-linux-arm64.zip` |
| **Linux** | 32-bit | `wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-linux-386.zip` |
| **macOS** | Intel | `curl -L -o mac.zip https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-darwin-amd64.zip` |
| **macOS** | Apple Silicon | `curl -L -o mac.zip https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-darwin-arm64.zip` |
| **Windows** | 64-bit | Download: [whatsapp-web-api-windows-amd64.exe.zip](https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-windows-amd64.exe.zip) |
| **Windows** | 32-bit | Download: [whatsapp-web-api-windows-386.exe.zip](https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-windows-386.exe.zip) |

2. **Extract and make executable**:
   ```bash
   # For Linux/macOS
   unzip whatsapp-web-api-linux-amd64.zip
   chmod +x whatsapp-web-api-linux-amd64

   # For Windows - just unzip the .exe file
   # No additional steps needed
   ```

3. **Set up environment**:
   ```bash
   # Create environment file
   nano .env
   ```

   **Add your database configuration**:
   ```env
   DATABASE_URL=postgres://username:password@localhost:5432/whatsapp_db
   PORT=8080
   WA_WEBHOOK_URL=https://your-webhook.com/webhook
   ```

4. **Run the API server**:
   ```bash
   # Linux/macOS
   ./whatsapp-web-api-linux-amd64

   # Windows
   whatsapp-web-api-windows-amd64.exe
   ```

5. **Pair with WhatsApp**:
   ```bash
   # Get QR code for pairing
   curl http://localhost:8080/pair
   ```

6. **Start sending messages**:
    ```bash
    # Send a message with attachment
    curl -X POST http://localhost:8080/send \
      -H "Content-Type: application/json" \
      -d '{
        "number":"1234567890",
        "message":"Hello from binary!",
        "attachments":[{
          "type":"image",
           "url":"https://picsum.photos/800/600",
          "caption":"Check this out"
        }]
      }'
    ```

### **Production Deployment with Binary**

#### **Manual Systemd Service Setup**:
```bash
# 1. Move binary to system location
sudo mv whatsapp-web-api-linux-amd64 /opt/whatsapp-api/whatsapp-web-api
sudo chmod +x /opt/whatsapp-api/whatsapp-web-api

# 2. Create service user
sudo useradd -r -s /bin/false whatsapp-api

# 3. Create environment file
sudo nano /opt/whatsapp-api/.env
# Add your DATABASE_URL and other vars

# 4. Create systemd service
sudo nano /etc/systemd/system/whatsapp-api.service
```

**Service file content**:
```ini
[Unit]
Description=WhatsApp Web API
After=network.target

[Service]
Type=simple
User=whatsapp-api
WorkingDirectory=/opt/whatsapp-api
ExecStart=/opt/whatsapp-api/whatsapp-web-api
Restart=always
RestartSec=10
EnvironmentFile=/opt/whatsapp-api/.env

[Install]
WantedBy=multi-user.target
```

```bash
# 5. Start and enable service
sudo systemctl daemon-reload
sudo systemctl start whatsapp-api
sudo systemctl enable whatsapp-api

# 6. Check status
sudo systemctl status whatsapp-api
```

#### **Automated Deployment** (included with binary):
```bash
# Download and run the automated deployment script
wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/whatsapp-web-api-linux-amd64.zip
unzip whatsapp-web-api-linux-amd64.zip
DATABASE_URL="postgres://user:pass@host:5432/db" sudo ./deploy-linux.sh
```

### **Binary Verification (Security)**

Always verify downloaded binaries:

```bash
# Download checksums
wget https://github.com/rizrmd/whatsapp-web-api/releases/download/v1.6.0/checksums.txt

# Verify your binary
sha256sum whatsapp-web-api-linux-amd64.zip
# Compare with the value in checksums.txt
```

## 💻 Usage Examples

### cURL Examples

**Check health status**:
```bash
curl http://localhost:8080/health
```

**Generate QR code**:
```bash
curl http://localhost:8080/pair
```

**Send message with attachments**:
```bash
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{
    "number": "1234567890",
    "message": "Hello from API!",
    "attachments": [
      {
        "type": "image",
      "url": "https://picsum.photos/800/600",
        "caption": "Check this out"
      }
    ]
  }'
```

**Send text-only message**:
```bash
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{
    "number": "1234567890",
    "message": "Hello from API!"
  }'
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

// Send message with attachments
async function sendMessage(number, message, attachments = []) {
  try {
    const response = await axios.post('http://localhost:8080/send', {
      number: number,
      message: message,
      attachments: attachments
    });
    console.log('Message sent:', response.data);
  } catch (error) {
    console.error('Error:', error.response.data);
  }
}

// Usage examples
sendMessage('1234567890', 'Hello from Node.js!');

// Send with image attachment
sendMessage('1234567890', 'Check this image!', [
  {
    type: 'image',
    url: 'https://picsum.photos/800/600',
    caption: 'Amazing photo'
  }
]);
```

### Python Example

```python
import requests
import json

def send_whatsapp_message(number, message, attachments=None):
    url = "http://localhost:8080/send"
    payload = {
        "number": number,
        "message": message
    }
    
    if attachments:
        payload["attachments"] = attachments

    response = requests.post(url, json=payload)
    return response.json()

# Usage examples
result = send_whatsapp_message("1234567890", "Hello from Python!")
print(result)

# Send with document attachment
result = send_whatsapp_message("1234567890", "Here's the document", [
    {
        "type": "document",
        "url": "https://example.com/document.pdf",
        "filename": "report.pdf"
    }
])
print(result)

# Access downloaded image (from webhook)
def download_image(image_filename, save_path):
    url = f"http://localhost:8080/images/{image_filename}"
    response = requests.get(url)
    
    if response.status_code == 200:
        with open(save_path, 'wb') as f:
            f.write(response.content)
        print(f"Image saved to {save_path}")
    else:
        print(f"Failed to download image: {response.status_code}")

# Example usage after receiving webhook with image URL
# download_image("ABC123.jpg", "downloaded_image.jpg")
```

## 🪝 Webhook Integration

When `WA_WEBHOOK_URL` is configured, incoming messages are sent as POST requests:

**Webhook Payload**:
```json
{
  "event": "message",
  "message": "Hello there!",
  "sender": "1234567890@s.whatsapp.net",
  "chat": "1234567890-1234567890@g.us",
  "time": "2025-10-25T16:07:24Z",
  "attachment": {
    "type": "image",
    "caption": "vacation photo",
    "mimetype": "image/jpeg",
    "file_length": 1024000,
    "width": 1920,
    "height": 1080,
    "url": "/images/ABC123.jpg"
  }
}
```

**Enhanced Attachment Support**:
- **Images**: Dimensions, file size, caption, and accessible URL
- **Documents**: Title, MIME type, file size, page count
- **Audio**: Duration, MIME type, file size
- **Video**: Dimensions, duration, caption, MIME type, file size
- **Stickers**: Dimensions, MIME type, file size
- **Contacts**: Display name, vCard data
- **Locations**: Name, address, coordinates

**Webhook Server Example (Node.js)**:
```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.post('/webhook', (req, res) => {
  const { event, message, sender, chat, time } = req.body;

  console.log(`Received ${event}: ${message} from ${sender}`);

  // Process the message...

  res.status(200).send('OK');
});

app.listen(3000, () => {
  console.log('Webhook server running on port 3000');
});
```

## 📚 Documentation

### Swagger/OpenAPI
- **Basic Info**: `GET /swagger`
- **Full Spec**: `GET /swagger.yaml`
- **Swagger UI**: Use [https://editor.swagger.io/](https://editor.swagger.io/) with the YAML file

### Complete API Documentation
See `swagger.yaml` for detailed OpenAPI 3.0 specification including:
- All endpoints and parameters
- Request/response schemas
- Error handling
- Authentication information

## 🐳 Docker Support

**Dockerfile**:
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o whatsapp-web-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/whatsapp-web-api .
COPY --from=builder /app/swagger.yaml .
CMD ["./whatsapp-web-api"]
EXPOSE 8080
```

**Build and run**:
```bash
docker build -t whatsapp-web-api .
docker run -p 8080:8080 --env-file .env whatsapp-web-api
```

## 🔧 Configuration

### Server Configuration
- **Port**: Set via `PORT` environment variable (default: 8080)
- **Database**: PostgreSQL connection via `DATABASE_URL`
- **Webhooks**: Incoming message forwarding via `WA_WEBHOOK_URL`

### Security Considerations
- 🔐 **Database Security**: Sessions stored securely in PostgreSQL
- 🛡️ **SSL/TLS**: Auto-configured SSL mode for database connections
- 🔑 **Environment Variables**: Sensitive data via env vars only
- 🌐 **Webhook Security**: Validate webhook requests at your endpoint

## 🚨 Error Handling

All API responses follow consistent format:

```json
{
  "success": false,
  "message": "Error description",
  "data": {}
}
```

**Common HTTP Status Codes**:
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `405` - Method Not Allowed
- `422` - Unprocessable Entity (not paired)
- `500` - Internal Server Error

## 🐛 Troubleshooting

### Database Issues
```bash
# Test database connection
psql $DATABASE_URL

# Check if database exists
\l

# Create tables if needed (handled automatically)
```

### Pairing Problems
- ✅ Enable multi-device in WhatsApp Settings
- ✅ Scan QR code within timeout (15 seconds)
- ✅ Ensure stable internet connection
- ✅ Check WhatsApp app is updated
- ✅ **Device Limits**: WhatsApp allows maximum 4 connected devices per account
- ✅ **Check Connected Devices**: Go to WhatsApp Settings > Linked Devices to see current connections
- ✅ **Remove Old Devices**: Disconnect unused devices from WhatsApp mobile app
- ✅ **Use New Endpoints**: 
  - `GET /devices` - Check current device status
  - `POST /disconnect` - Clean up problematic sessions
- ✅ **Session Cleanup**: If pairing fails, try disconnecting first, then pair again

### Device Limit Issues
WhatsApp has a **4-device limit** per account. If you're experiencing pairing issues:

1. **Check Connected Devices**:
   ```bash
   curl http://localhost:8080/devices
   ```

2. **Remove Unused Devices**:
   - Open WhatsApp mobile app
   - Go to Settings > Linked Devices
   - Tap and hold on unused devices
   - Select "Log out"

3. **Clean Current Session**:
   ```bash
   curl -X POST http://localhost:8080/disconnect
   ```

4. **Try Pairing Again**:
   ```bash
   curl http://localhost:8080/pair
   ```

### Connection Issues
- ✅ **Check Device Status**: Use `GET /health` and `GET /devices` endpoints
- ✅ **Network Stability**: Ensure stable internet connection on both server and phone
- ✅ **Session Corruption**: If connection fails, use `POST /disconnect` to clear session
- ✅ **WhatsApp Updates**: Keep WhatsApp mobile app updated
- ✅ **Multi-Device**: Ensure multi-device is enabled in WhatsApp settings

### Message Sending Issues
- ✅ Verify pairing with `/health` endpoint
- ✅ Check phone number format: `1234567890` (no '+')
- ✅ Ensure message length < 4096 characters
- ✅ Check webhook status if messages aren't being received
- ✅ **Important**: Attachment URLs must be publicly accessible HTTP/HTTPS links (not base64 data)
- ✅ Test attachment URLs in browser to ensure they're accessible
- ✅ For single image + text: Text becomes image caption (combined message)
- ✅ For multiple images: Each image sends as separate message

### Image Access Issues
- ✅ Images are automatically downloaded to `downloads/` directory
- ✅ Access images via `/images/{filename}` endpoint
- ✅ Check if image file exists in downloads directory
- ✅ Verify webhook payload contains correct image URL
- ✅ Images are served with appropriate MIME types and caching headers

### Server Issues
```bash
# Check if port is available
netstat -tulpn | grep 8080

# Run with verbose logging
GODEBUG=debug ./whatsapp-web-api

# Check database logs
tail -f /var/log/postgresql/postgresql-*.log
```

## 📈 Production Deployment

### Environment Setup
```bash
# Production environment variables
export DATABASE_URL="postgres://user:pass@prod-db:5432/whatsapp"
export PORT="8080"
export WA_WEBHOOK_URL="https://webhook.example.com/whatsapp"
```

### Systemd Service
```ini
# /etc/systemd/system/whatsapp-api.service
[Unit]
Description=WhatsApp Web API
After=network.target

[Service]
Type=simple
User=whatsapp
WorkingDirectory=/opt/whatsapp-api
ExecStart=/opt/whatsapp-api/whatsapp-web-api
Restart=always
RestartSec=10
Environment=DATABASE_URL=postgres://user:pass@localhost:5432/whatsapp

[Install]
WantedBy=multi-user.target
```

### Monitoring
- Monitor `/health` endpoint for availability
- Track webhook delivery status
- Monitor database connection pools
- Set up alerts for pairing failures

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 Documentation: Check this README and `swagger.yaml`
- 🐛 Issues: Create GitHub issues for bugs
- 💬 Discussions: Use GitHub Discussions for questions
- 📧 Email: support@example.com (if provided)

---

**Made with ❤️ for WhatsApp automation**