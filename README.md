# NetPeek

Network monitoring command-line tool built with Go.

## Features

- **IP Lookup**: Check public IP address from multiple sources (both domestic and international)
- **WebSocket Test**: Test WebSocket connections with support for ws:// and wss:// protocols

## Installation

```bash
# Clone the repository
git clone https://github.com/zhuinfra/netpeek.git
cd netpeek

# Build the project
go build

# Run the application
./netpeek --help
```

## Usage

### IP Lookup
```bash
./netpeek ip
```

This will display your public IP address from multiple sources:
- Domestic: Bilibili API and Tencent API
- International: Cloudflare API and IPinfo.io API

### WebSocket Test
```bash
./netpeek ws <websocket-url>

# Example
./netpeek ws wss://echo.websocket.org
```

## Project Structure

```
netpeek/
├── main.go              # Main application entry point
├── ip/                  # IP lookup functionality
│   └── ip.go
├── websocket/           # WebSocket testing functionality
│   └── ws.go
├── utils/               # Utility functions (future use)
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── README.md            # This file
```

## Adding New Features

To add new features to NetPeek:

1. Create a new directory at the root level for your feature (similar to `ip/` and `websocket/`)
2. Implement the feature in that directory
3. Add a new command case in `main.go`
4. Update the help information in the printHelp() function

## License

MIT
