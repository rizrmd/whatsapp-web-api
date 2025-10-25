#!/bin/bash

# WhatsApp Web API Linux Deployment Script
# This script helps deploy the WhatsApp Web API on Linux systems

set -e

# Configuration
BINARY_NAME="whatsapp-web-api"
SERVICE_NAME="whatsapp-api"
INSTALL_DIR="/opt/whatsapp-api"
SERVICE_USER="whatsapp"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   print_error "This script must be run as root"
   exit 1
fi

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            BINARY_ARCH="amd64"
            ;;
        aarch64|arm64)
            BINARY_ARCH="arm64"
            ;;
        i386|i686)
            BINARY_ARCH="386"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    print_status "Detected architecture: $ARCH ($BINARY_ARCH)"
}

# Check for required environment variables
check_env() {
    print_status "Checking environment variables..."

    if [[ -z "$DATABASE_URL" ]]; then
        print_error "DATABASE_URL environment variable is required"
        exit 1
    fi

    print_success "Environment variables check passed"
}

# Create service user
create_user() {
    print_status "Creating service user..."

    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd -r -s /bin/false -d $INSTALL_DIR $SERVICE_USER
        print_success "Created service user: $SERVICE_USER"
    else
        print_warning "Service user $SERVICE_USER already exists"
    fi
}

# Create installation directory
create_directory() {
    print_status "Creating installation directory..."

    mkdir -p $INSTALL_DIR
    chown $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
    chmod 750 $INSTALL_DIR
    print_success "Created directory: $INSTALL_DIR"
}

# Install binary
install_binary() {
    print_status "Installing binary..."

    BINARY_FILE="${BINARY_NAME}-linux-${BINARY_ARCH}"

    if [[ ! -f "$BINARY_FILE" ]]; then
        print_error "Binary file $BINARY_FILE not found"
        exit 1
    fi

    cp "$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
    chown $SERVICE_USER:$SERVICE_USER "$INSTALL_DIR/$BINARY_NAME"
    chmod 755 "$INSTALL_DIR/$BINARY_NAME"

    print_success "Installed binary: $INSTALL_DIR/$BINARY_NAME"
}

# Install swagger documentation
install_docs() {
    print_status "Installing documentation..."

    if [[ -f "swagger.yaml" ]]; then
        cp swagger.yaml $INSTALL_DIR/
        chown $SERVICE_USER:$SERVICE_USER "$INSTALL_DIR/swagger.yaml"
        print_success "Installed documentation: $INSTALL_DIR/swagger.yaml"
    fi
}

# Create systemd service
create_service() {
    print_status "Creating systemd service..."

    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=WhatsApp Web API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=always
RestartSec=10
Environment=DATABASE_URL=$DATABASE_URL
EnvironmentFile=-$INSTALL_DIR/.env

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    print_success "Created systemd service: $SERVICE_NAME"
}

# Setup environment file
setup_env() {
    print_status "Setting up environment file..."

    cat > $INSTALL_DIR/.env << EOF
# WhatsApp Web API Environment Variables
DATABASE_URL=$DATABASE_URL
PORT=${PORT:-8080}

# Optional: Webhook URL
# WA_WEBHOOK_URL=https://your-webhook-endpoint.com/webhook
EOF

    chown $SERVICE_USER:$SERVICE_USER $INSTALL_DIR/.env
    chmod 640 $INSTALL_DIR/.env

    print_success "Created environment file: $INSTALL_DIR/.env"
}

# Setup log rotation
setup_logrotate() {
    print_status "Setting up log rotation..."

    cat > /etc/logrotate.d/${SERVICE_NAME} << EOF
$INSTALL_DIR/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 $SERVICE_USER $SERVICE_USER
    postrotate
        systemctl reload ${SERVICE_NAME}
    endscript
}
EOF

    print_success "Set up log rotation"
}

# Main installation
main() {
    print_status "Starting WhatsApp Web API installation..."

    detect_arch
    check_env
    create_user
    create_directory
    install_binary
    install_docs
    setup_env
    create_service
    setup_logrotate

    print_success "Installation completed successfully!"
    print_status "To start the service:"
    echo -e "${GREEN}sudo systemctl start $SERVICE_NAME${NC}"
    echo -e "${GREEN}sudo systemctl enable $SERVICE_NAME${NC}"
    print_status "To check service status:"
    echo -e "${GREEN}sudo systemctl status $SERVICE_NAME${NC}"
    print_status "To view logs:"
    echo -e "${GREEN}sudo journalctl -u $SERVICE_NAME -f${NC}"
}

# Show usage
show_usage() {
    echo "Usage: $0 [ENVIRONMENT_VARS]"
    echo ""
    echo "Required environment variables:"
    echo "  DATABASE_URL - PostgreSQL connection string"
    echo ""
    echo "Optional environment variables:"
    echo "  PORT - Server port (default: 8080)"
    echo "  WA_WEBHOOK_URL - Webhook endpoint for incoming messages"
    echo ""
    echo "Example:"
    echo "  DATABASE_URL=postgres://user:pass@localhost:5432/whatsapp $0"
    echo "  DATABASE_URL=postgres://user:pass@localhost:5432/whatsapp PORT=3000 $0"
}

# Check for help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

# Run main function
main