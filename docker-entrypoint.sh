#!/bin/sh
set -e

# Detect if chromium or chromium-browser is available
if command -v chromium >/dev/null 2>&1 || command -v chromium-browser >/dev/null 2>&1; then
    echo "Chromium already installed"
else
    echo "Installing Chromium..."

    # Detect package manager and install accordingly
    if command -v apk >/dev/null 2>&1; then
        # Alpine Linux
        echo "Detected Alpine Linux, using apk..."
        apk add --no-cache \
            chromium \
            chromium-chromedriver \
            nss \
            freetype \
            harfbuzz \
            ca-certificates \
            ttf-freefont \
            font-noto-emoji \
            su-exec

        # Set Chrome binary path for Alpine
        export CHROME_BIN=/usr/bin/chromium
        echo "Chromium installed successfully (Alpine)"

    elif command -v apt-get >/dev/null 2>&1; then
        # Debian/Ubuntu
        echo "Detected Debian/Ubuntu, using apt-get..."
        apt-get update -qq
        apt-get install -y -qq \
            chromium \
            chromium-driver \
            fonts-liberation \
            libnss3 \
            libatk-bridge2.0-0 \
            libxcomposite1 \
            libxdamage1 \
            libxrandr2 \
            libgbm1 \
            libasound2 \
            gosu \
            > /dev/null 2>&1
        apt-get clean
        rm -rf /var/lib/apt/lists/*

        export CHROME_BIN=/usr/bin/chromium-browser
        echo "Chromium installed successfully (Debian)"

    else
        echo "Error: Unable to detect package manager (apk or apt-get)"
        exit 1
    fi
fi

# Verify Chromium is available and show version
if command -v chromium >/dev/null 2>&1; then
    echo "Chromium version: $(chromium --version)"
    export CHROME_BIN=/usr/bin/chromium
elif command -v chromium-browser >/dev/null 2>&1; then
    echo "Chromium version: $(chromium-browser --version)"
    export CHROME_BIN=/usr/bin/chromium-browser
else
    echo "Warning: Chromium not found after installation!"
fi

# Switch to grafana user and start Grafana
echo "Starting Grafana..."

# Check if we're already running as grafana user
if [ "$(id -u)" = "0" ]; then
    # Running as root, need to switch to grafana user
    if command -v su-exec >/dev/null 2>&1; then
        # Alpine: use su-exec
        exec su-exec grafana /run.sh "$@"
    elif command -v gosu >/dev/null 2>&1; then
        # Debian: use gosu
        exec gosu grafana /run.sh "$@"
    else
        # Fallback: use su
        exec su -s /bin/sh grafana -c "exec /run.sh $*"
    fi
else
    # Already running as non-root user
    exec /run.sh "$@"
fi
