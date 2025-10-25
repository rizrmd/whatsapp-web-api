# WhatsApp Web API

🚀 A production-ready REST API for WhatsApp Web integration that enables programmatic sending of messages, QR code pairing, and webhook-based message reception.

## ✨ Features

- **📱 QR Code Pairing**: Generate QR codes to pair WhatsApp accounts
- **💬 Message Sending**: Send text messages to any WhatsApp number via REST API
- **🪝 Webhook Support**: Receive incoming messages via HTTP webhooks
- **🗄️ PostgreSQL Storage**: Secure WhatsApp session persistence in PostgreSQL
- **🔒 Auto SSL Handling**: Automatically configures PostgreSQL SSL mode
- **📚 Complete Documentation**: Full OpenAPI 3.0 / Swagger specification
- **🔄 Session Management**: Automatic reconnection and session handling
- **⚡ High Performance**: Concurrent message handling and graceful shutdown

## 📋 Requirements

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

1. **Clone and setup**:
   ```bash
   git clone <your-repo>
   cd whatsapp-web-api
   cp .env.example .env
   # Edit .env with your credentials
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build and run**:
   ```bash
   go build -o whatsapp-web-api .
   ./whatsapp-web-api
   ```

4. **Verify installation**:
   ```bash
   curl http://localhost:8080/health
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

Send a text message to a WhatsApp number.

**Request Body**:
```json
{
  "number": "1234567890",
  "message": "Hello from WhatsApp API!"
}
```

**Parameters**:
- `number` (string, required): Phone number with country code (no '+' prefix)
- `message` (string, required): Message text (max 4096 characters)

**Response**:
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "number": "1234567890",
    "message": "Hello from WhatsApp API!"
  }
}
```

### 4. API Documentation
```http
GET /swagger
GET /swagger.yaml
```

Access API documentation and OpenAPI specification.

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

**Send message**:
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

// Send message
async function sendMessage(number, message) {
  try {
    const response = await axios.post('http://localhost:8080/send', {
      number: number,
      message: message
    });
    console.log('Message sent:', response.data);
  } catch (error) {
    console.error('Error:', error.response.data);
  }
}

// Usage
sendMessage('1234567890', 'Hello from Node.js!');
```

### Python Example

```python
import requests
import json

def send_whatsapp_message(number, message):
    url = "http://localhost:8080/send"
    payload = {
        "number": number,
        "message": message
    }

    response = requests.post(url, json=payload)
    return response.json()

# Usage
result = send_whatsapp_message("1234567890", "Hello from Python!")
print(result)
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
  "time": "2025-10-25T16:07:24Z"
}
```

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
- ✅ Scan QR code within timeout (60 seconds)
- ✅ Ensure stable internet connection
- ✅ Check WhatsApp app is updated

### Message Sending Issues
- ✅ Verify pairing with `/health` endpoint
- ✅ Check phone number format: `1234567890` (no '+')
- ✅ Ensure message length < 4096 characters
- ✅ Check webhook status if messages aren't being received

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