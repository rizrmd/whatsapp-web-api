# WhatsApp Web API - Troubleshooting Guide

## 🚨 Common Issues & Solutions

### 1. QR Code Pairing Issues

#### Problem: QR code generation timeout
**Error**: `QR code generation timeout - please try again`

**Solutions**:
1. **Check Internet Connection**: Ensure stable internet on both server and phone
2. **Extended Timeout**: The app now waits 15 seconds (increased from 10)
3. **Session Cleanup**: Run `POST /disconnect` before trying again
4. **Device Limits**: Check if you have too many devices connected

```bash
# Clean up existing session
curl -X POST http://localhost:8080/disconnect

# Try pairing again
curl http://localhost:8080/pair
```

#### Problem: "Device limit exceeded" error
**Error**: `Device limit exceeded on WhatsApp account`

**Solutions**:
1. **Check Connected Devices**: Go to WhatsApp mobile app > Settings > Linked Devices
2. **Remove Unused Devices**: Tap and hold on old devices, select "Log out"
3. **WhatsApp Limit**: Maximum 4 devices per account
4. **Clean Current Session**: Use `POST /disconnect` endpoint

```bash
# Check current device status
curl http://localhost:8080/devices

# Clean up and start fresh
curl -X POST http://localhost:8080/disconnect
```

#### Problem: "Multi-device not enabled" error
**Error**: `QR code was scanned but multidevice is not enabled on the phone`

**Solutions**:
1. **Enable Multi-Device**: 
   - Open WhatsApp mobile app
   - Go to Settings > Linked Devices
   - Enable "Multi-device" if not already enabled
2. **Update WhatsApp**: Ensure you have the latest version
3. **Restart WhatsApp**: Force close and reopen WhatsApp mobile app

### 2. Connection Issues

#### Problem: Failed to connect to existing session
**Error**: `Failed to connect to existing session`

**Solutions**:
1. **Session Corruption**: The existing session may be corrupted
2. **Device Disconnected**: Device may have been disconnected from mobile app
3. **Network Issues**: Check internet connectivity
4. **Clean Session**: Use `POST /disconnect` to clear and start fresh

```bash
# Check current status
curl http://localhost:8080/health

# Clean up problematic session
curl -X POST http://localhost:8080/disconnect

# Create new session
curl http://localhost:8080/pair
```

#### Problem: Stream errors during operation
**Error**: `Stream error occurred`

**Solutions**:
1. **Network Stability**: Ensure stable internet connection
2. **Session Health**: Check if session is still valid
3. **Device Limits**: May indicate device limit issues
4. **Restart Session**: Disconnect and pair again

### 3. Message Sending Issues

#### Problem: "Not paired with WhatsApp" error
**Error**: `Not paired with WhatsApp. Please use /pair endpoint first`

**Solutions**:
1. **Check Pairing Status**: Verify device is paired and connected
2. **Health Check**: Use `GET /health` endpoint
3. **Re-pair if Needed**: Use `GET /pair` to create new session

```bash
# Check status
curl http://localhost:8080/health

# If not paired, create new session
curl http://localhost:8080/pair
```

#### Problem: Invalid phone number format
**Error**: `Invalid phone number`

**Solutions**:
1. **Format**: Use country code without '+' (e.g., `1234567890`)
2. **Country Code**: Include correct country code
3. **No Special Characters**: Only numbers allowed

```bash
# Correct format
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{"number":"1234567890","message":"Hello!"}'

# Incorrect format (will fail)
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{"number":"+1234567890","message":"Hello!"}'
```

### 4. Database Issues

#### Problem: Database connection failed
**Error**: `Failed to create database container`

**Solutions**:
1. **Check DATABASE_URL**: Verify connection string is correct
2. **Database Running**: Ensure PostgreSQL server is running
3. **Network Access**: Check firewall and network connectivity
4. **Credentials**: Verify username/password are correct

```bash
# Test database connection
psql $DATABASE_URL

# Check if database exists
\l
```

#### Problem: SSL mode issues
**Error**: SSL-related connection errors

**Solutions**:
1. **Auto SSL Mode**: The app automatically adds `sslmode=disable` if not specified
2. **Manual SSL**: You can specify SSL mode in DATABASE_URL
3. **Local Development**: Use `sslmode=disable` for local PostgreSQL

```env
# Example DATABASE_URL with SSL mode
DATABASE_URL=postgres://user:pass@localhost:5432/db?sslmode=disable
```

### 5. Performance & Resource Issues

#### Problem: Port already in use
**Error**: `listen tcp :8080: bind: address already in use`

**Solutions**:
1. **Change Port**: Set different PORT environment variable
2. **Kill Process**: Find and kill process using port 8080
3. **Restart Service**: If running as service, restart properly

```bash
# Find process using port 8080
lsof -i :8080

# Kill process (replace PID)
kill -9 <PID>

# Or use different port
PORT=8081 ./wameow
```

#### Problem: Memory usage high
**Symptoms**: High memory consumption over time

**Solutions**:
1. **Session Cleanup**: Regular cleanup of old sessions
2. **Image Downloads**: Monitor downloads directory size
3. **Log Rotation**: Implement log rotation for production
4. **Resource Limits**: Set appropriate resource limits

```bash
# Check downloads directory size
du -sh downloads/

# Clean old images (older than 7 days)
find downloads/ -name "*.jpg" -mtime +7 -delete
```

## 🔧 Advanced Troubleshooting

### Debug Mode
Enable verbose logging for debugging:

```bash
# Run with debug logging
GODEBUG=debug ./wameow
```

### Session Management
Advanced session management commands:

```bash
# Check device information
curl http://localhost:8080/devices

# Force disconnect and cleanup
curl -X POST http://localhost:8080/disconnect

# Check overall health
curl http://localhost:8080/health
```

### WhatsApp Device Management
1. **Open WhatsApp Mobile App**
2. **Go to Settings > Linked Devices**
3. **View Connected Devices**: See all currently connected devices
4. **Remove Devices**: Tap and hold, select "Log out"
5. **Device Limit**: Maximum 4 devices per account

### Log Analysis
Common log patterns and their meanings:

```
🟢 Connected to WhatsApp!           # Successful connection
🔴 Disconnected from WhatsApp       # Lost connection
🎉 Successfully paired!             # QR code scanned successfully
⏰ QR code pairing timed out        # QR not scanned in time
🚫 Device limit exceeded           # Too many devices connected
📱 Multi-device not enabled         # Enable in WhatsApp settings
```

## 📞 Getting Help

### Check Application Status
```bash
# Health check
curl http://localhost:8080/health

# Device info
curl http://localhost:8080/devices

# Test all endpoints
./test_endpoints.sh
```

### Common Debugging Steps
1. **Check Logs**: Look for error messages in application logs
2. **Verify Environment**: Check all environment variables
3. **Test Database**: Verify database connectivity
4. **Network Check**: Ensure internet connectivity
5. **WhatsApp Status**: Check if WhatsApp is working on mobile

### Environment Variables Checklist
```env
# Required
DATABASE_URL=postgres://user:pass@host:5432/db

# Optional
PORT=8080
WA_WEBHOOK_URL=https://your-webhook.com/webhook
```

### Production Considerations
- **Monitoring**: Monitor `/health` endpoint
- **Logging**: Implement log rotation
- **Backups**: Regular database backups
- **Security**: Use HTTPS in production
- **Resource Limits**: Set appropriate memory/CPU limits

---

## 🆘 Still Having Issues?

If you're still experiencing problems:

1. **Check the logs** for detailed error messages
2. **Try the basic troubleshooting steps** above
3. **Use the test script** to verify endpoints
4. **Check GitHub Issues** for known problems
5. **Create a new issue** with detailed information

**When creating an issue, include**:
- Error messages (full logs)
- Environment details (OS, Go version)
- WhatsApp version
- Steps to reproduce
- Expected vs actual behavior