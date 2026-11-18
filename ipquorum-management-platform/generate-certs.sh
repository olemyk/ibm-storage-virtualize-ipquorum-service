#!/usr/bin/env bash
set -euo pipefail

# Generate self-signed TLS certificates for IPQuorum Manager Server
# Usage: ./generate-certs.sh [output_dir]

OUTPUT_DIR="${1:-./tls}"
CERT_DAYS=365
COUNTRY="NO"
STATE="Oslo"
CITY="Oslo"
ORG="IPQuorum"
OU="Development"
CN="localhost"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔐 Generating Self-Signed TLS Certificates"
echo "=========================================="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"
echo "Output directory: $OUTPUT_DIR"

# Generate private key
echo -e "${YELLOW}Generating private key...${NC}"
openssl genrsa -out "$OUTPUT_DIR/server.key" 2048 2>/dev/null
chmod 600 "$OUTPUT_DIR/server.key"
echo -e "${GREEN}✓ Private key generated${NC}"

# Generate certificate signing request (CSR)
echo -e "${YELLOW}Generating certificate signing request...${NC}"
openssl req -new -key "$OUTPUT_DIR/server.key" \
  -out "$OUTPUT_DIR/server.csr" \
  -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=$CN" \
  2>/dev/null
echo -e "${GREEN}✓ CSR generated${NC}"

# Generate self-signed certificate
echo -e "${YELLOW}Generating self-signed certificate (valid for $CERT_DAYS days)...${NC}"
openssl x509 -req -days $CERT_DAYS \
  -in "$OUTPUT_DIR/server.csr" \
  -signkey "$OUTPUT_DIR/server.key" \
  -out "$OUTPUT_DIR/server.crt" \
  -extfile <(printf "subjectAltName=DNS:localhost,DNS:ipquorum-manager,IP:127.0.0.1") \
  2>/dev/null
chmod 644 "$OUTPUT_DIR/server.crt"
echo -e "${GREEN}✓ Certificate generated${NC}"

# Clean up CSR
rm -f "$OUTPUT_DIR/server.csr"

# Display certificate info
echo ""
echo "=========================================="
echo -e "${GREEN}✅ Certificates generated successfully!${NC}"
echo ""
echo "Files created:"
echo "  - $OUTPUT_DIR/server.key (private key)"
echo "  - $OUTPUT_DIR/server.crt (certificate)"
echo ""
echo "Certificate details:"
openssl x509 -in "$OUTPUT_DIR/server.crt" -noout -subject -dates 2>/dev/null
echo ""
echo "To use with IPQuorum Manager:"
echo "  1. Mount certificates: -v \$(pwd)/tls:/etc/ipquorum-platform/tls:ro"
echo "  2. Or set environment variables:"
echo "     -e TLS_CERT_FILE=/path/to/server.crt"
echo "     -e TLS_KEY_FILE=/path/to/server.key"
echo ""
echo -e "${YELLOW}⚠️  Note: Self-signed certificates will show security warnings in browsers.${NC}"
echo "    For production, use certificates from a trusted CA."
echo ""

# Made with Bob