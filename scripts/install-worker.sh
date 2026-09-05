#!/usr/bin/env bash
# ==============================================================================
# ANARVA Cloud Compute Worker Host Installer & Security Hardening Script
# Target OS: Ubuntu 24.04 LTS / Debian 12 / Linux Kernel >= 5.15
# ==============================================================================

set -euo pipefail

echo "[ANARVA Worker Installer] Starting production worker setup..."

if [ "$EUID" -ne 0 ]; then
  echo "FATAL: Installer must be run as root." >&2
  exit 1
fi

# 1. Create Dedicated Unprivileged User & Group
if ! id -u anarva-worker >/dev/null 2>&1; then
  echo "[1/6] Creating system user 'anarva-worker'..."
  useradd --system --shell /sbin/nologin --comment "ANARVA Compute Worker Data Plane User" anarva-worker
fi

if getent group docker >/dev/null 2>&1; then
  echo "[1/6] Adding 'anarva-worker' to 'docker' group..."
  usermod -aG docker anarva-worker
else
  echo "FATAL: 'docker' group does not exist. Ensure Docker Engine is installed." >&2
  exit 1
fi

# 2. Setup Directory Structure & Permissions
echo "[2/6] Setting up restricted filesystem directories..."
mkdir -p /etc/anarva/certs
mkdir -p /var/lib/anarva-worker

chown -R anarva-worker:docker /etc/anarva
chown -R anarva-worker:docker /var/lib/anarva-worker

chmod 0750 /etc/anarva
chmod 0700 /etc/anarva/certs
chmod 0750 /var/lib/anarva-worker

# 3. Configure Host Firewall Cloud Metadata Endpoint Protection (169.254.169.254)
echo "[3/7] Configuring Docker metadata protection..."

if ! command -v iptables >/dev/null 2>&1; then
  echo "FATAL: iptables is required for production metadata protection." >&2
  exit 1
fi

# Docker invokes DOCKER-USER before its own forwarding rules. Keep ANARVA
# policy in this administrator-controlled chain rather than modifying Docker's
# managed FORWARD/DOCKER-FORWARD chains directly.
if ! iptables -C DOCKER-USER -d 169.254.169.254/32 -j DROP 2>/dev/null; then
  iptables -I DOCKER-USER 1 -d 169.254.169.254/32 -j DROP
fi

if ! iptables -C DOCKER-USER -d 169.254.169.254/32 -j DROP 2>/dev/null; then
  echo "FATAL: Failed to establish cloud metadata protection." >&2
  exit 1
fi

echo "[Firewall] DOCKER-USER metadata DROP rule verified."

# 4. Install ANARVA Firewall Helper
echo "[4/7] Installing ANARVA firewall helper..."

if [ -f "scripts/anarva-worker-firewall" ]; then
  install -o root -g root -m 0755 \
    scripts/anarva-worker-firewall \
    /usr/local/sbin/anarva-worker-firewall
else
  echo "FATAL: scripts/anarva-worker-firewall not found." >&2
  exit 1
fi

if [ ! -x /usr/local/sbin/anarva-worker-firewall ]; then
  echo "FATAL: Firewall helper installation failed." >&2
  exit 1
fi

# 5. Install Worker Executable Binary
if [ -f "bin/worker" ]; then
  echo "[5/7] Installing bin/worker binary to /usr/local/bin/worker..."
  cp bin/worker /usr/local/bin/worker
  chown root:root /usr/local/bin/worker
  chmod 0755 /usr/local/bin/worker
elif [ -f "/usr/local/bin/worker" ]; then
  echo "[5/7] Existing /usr/local/bin/worker binary detected."
else
  echo "WARNING: bin/worker binary not found in current directory. Please compile with 'go build -o bin/worker ./cmd/worker'."
fi

# 6. Install Systemd Service Units
echo "[6/7] Installing systemd service units..."

if [ ! -f "scripts/anarva-worker-firewall.service" ]; then
  echo "FATAL: scripts/anarva-worker-firewall.service not found." >&2
  exit 1
fi

if [ ! -f "scripts/anarva-worker.service" ]; then
  echo "FATAL: scripts/anarva-worker.service not found." >&2
  exit 1
fi

install -o root -g root -m 0644 \
  scripts/anarva-worker-firewall.service \
  /etc/systemd/system/anarva-worker-firewall.service

install -o root -g root -m 0644 \
  scripts/anarva-worker.service \
  /etc/systemd/system/anarva-worker.service

systemctl daemon-reload

# 6. Validate Environment Configuration & Fail Closed
echo "[7/7] Validating production environment configuration..."
ENV_FILE="/etc/anarva/worker.env"

if [ ! -f "$ENV_FILE" ]; then
  echo "[ANARVA Worker Installer] Creating template configuration file at $ENV_FILE..."
  cat <<EOF > "$ENV_FILE"
# ANARVA Worker Production Environment Configuration
WORKER_LISTEN_ADDR=:8443
WORKER_DB_PATH=/var/lib/anarva-worker/worker_metadata.json
ANARVA_ENV=production
WORKER_REQUIRE_CLIENT_AUTH=true

# SECURITY CONFIGURATION (REQUIRED IN PRODUCTION)
# COMPUTE_WORKER_TOKEN=
# COMPUTE_WORKER_CA_CERT=/etc/anarva/certs/ca.crt
# COMPUTE_WORKER_SERVER_CERT=/etc/anarva/certs/server.crt
# COMPUTE_WORKER_SERVER_KEY=/etc/anarva/certs/server.key
EOF
  chown anarva-worker:docker "$ENV_FILE"
  chmod 0600 "$ENV_FILE"
  echo "IMPORTANT: Fill in /etc/anarva/worker.env before starting the service."
fi

echo "[ANARVA Worker Installer] Installation completed cleanly."
echo "To enable and start the services:"
echo "  sudo systemctl enable --now anarva-worker-firewall"
echo "  sudo systemctl enable --now anarva-worker"
