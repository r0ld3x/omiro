# 🎥 Omiro - Random Video Chat Platform

A modern, real-time random video chat application similar to Omegle, built with Go and WebRTC. Connect with strangers worldwide through video, audio, and text chat.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![WebRTC](https://img.shields.io/badge/WebRTC-Enabled-brightgreen)
![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## ✨ Features

- 🎥 **Real-time Video & Audio Chat** - HD quality WebRTC peer-to-peer connections
- 💬 **Text Messaging** - Instant text chat alongside video
- 🔀 **Random Matching** - Smart queue-based matchmaking system
- ⏭️ **Next Person** - Skip to next match seamlessly (like Omegle)
- 🔒 **Session Management** - Secure token-based authentication
- 🚫 **Rate Limiting** - Protection against abuse and spam
- 🌐 **IP Banning** - Admin tools for moderation
- 📊 **Redis Backend** - Fast, scalable data storage
- 🎨 **Modern UI** - Beautiful, responsive design
- 🔄 **Auto-reconnect** - Automatic matchmaking on disconnect
- 🌍 **TURN/STUN Support** - Works behind firewalls and NAT

## 🏗️ Architecture

```
┌─────────────┐         WebSocket          ┌──────────────┐
│   Browser   │ ◄─────────────────────────► │   Go Server  │
│   Client    │         (Session Token)     │              │
└─────────────┘                              └──────┬───────┘
       │                                            │
       │ WebRTC (P2P)                              │
       │ Video/Audio                                │
       ▼                                            ▼
┌─────────────┐                            ┌──────────────┐
│   Browser   │                            │    Redis     │
│   Client    │                            │   Database   │
└─────────────┘                            └──────────────┘
```

### Components:

- **Frontend**: Vanilla JavaScript with WebRTC API
- **Backend**: Go with gorilla/websocket
- **Database**: Redis for session management, queue, and chat history
- **Signaling**: WebSocket for WebRTC negotiation
- **Media**: P2P WebRTC connections with TURN/STUN fallback

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21 or higher
- **Redis** 7.0 or higher
- Modern web browser with WebRTC support

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/r0ld3x/omiro.git
   cd omiro
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Start Redis**

   ```bash
   redis-server
   ```

4. **Configure environment** (optional)

   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

5. **Run the server**

   ```bash
   go run .
   ```

6. **Open in browser**
   ```
   http://localhost:8080/index.html
   ```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```env
# Server Configuration
PORT=8080
HOST=localhost

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Security
SESSION_SECRET=your-secret-key-here
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW=60

# WebRTC
STUN_SERVER=stun:stun.l.google.com:19302
TURN_SERVER=turn:openrelay.metered.ca:80
TURN_USERNAME=openrelayproject
TURN_PASSWORD=openrelayproject
```

### Redis Configuration

The application uses Redis for:

- Session token management
- Matchmaking queue
- Rate limiting
- IP ban list
- Active connection tracking
- Chat message history

## 📡 API Reference

### WebSocket Endpoint

**Connect:** `ws://localhost:8080/ws?token={session_token}`

### Session Management

#### Get Session Token

```http
GET /session/new
```

**Response:**

```json
{
  "token": "uuid:timestamp:signature"
}
```

### WebSocket Operations

#### Join Matchmaking Queue

```json
{
  "op": "join_queue"
}
```

#### Send Chat Message

```json
{
  "op": "chat",
  "data": {
    "message": "Hello!"
  }
}
```

#### Find Next Person

```json
{
  "op": "next"
}
```

#### WebRTC Signaling

**Send Offer:**

```json
{
  "op": "webrtc_offer",
  "data": {
    "sdp": "v=0\r\no=- ..."
  }
}
```

**Send Answer:**

```json
{
  "op": "webrtc_answer",
  "data": {
    "sdp": "v=0\r\no=- ..."
  }
}
```

**Send ICE Candidate:**

```json
{
  "op": "ice_candidate",
  "data": {
    "candidate": {
      "candidate": "candidate:...",
      "sdpMLineIndex": 0,
      "sdpMid": "0"
    }
  }
}
```

### Server Messages

#### Match Found

```json
{
  "op": "match_found",
  "partner": "partner-uuid",
  "should_call": true
}
```

#### Partner Disconnected

```json
{
  "op": "partner_disconnected"
}
```

## 🛠️ Technology Stack

### Backend

- **Go** - High-performance server
- **gorilla/websocket** - WebSocket implementation
- **redis/go-redis** - Redis client
- **google/uuid** - UUID generation

### Frontend

- **Vanilla JavaScript** - No framework dependencies
- **WebRTC API** - Peer-to-peer connections
- **WebSocket API** - Real-time communication
- **Modern CSS** - Responsive design

### Infrastructure

- **Redis** - Session & data storage
- **STUN/TURN** - NAT traversal
- **Ngrok/Cloudflare** - Deployment options

## 📁 Project Structure

```
omiro/
├── main.go                 # Application entry point
├── handle_websocket.go     # WebSocket connection handler
├── incoming.go             # Message routing
├── join_queue.go           # Matchmaking logic
├── handle_webrtc.go        # WebRTC signaling
├── chat.go                 # Chat message handling
├── client.go               # Client struct & methods
├── helper/
│   └── helper.go          # Utility functions
├── middleware/
│   ├── is_allowed.go      # Rate limiting & bans
│   └── session_token.go   # Session management
├── redis/
│   ├── client.go          # Redis connection
│   ├── operations.go      # Redis operations
│   ├── chat.go            # Chat storage
│   └── ips.go             # IP management
├── index.html             # Frontend application
└── go.mod                 # Go dependencies
```

## 🔒 Security Features

- ✅ **Session Tokens** - Cryptographically signed tokens
- ✅ **Rate Limiting** - Per-IP connection limits
- ✅ **IP Banning** - Persistent ban storage in Redis
- ✅ **Origin Checking** - WebSocket origin validation
- ✅ **Real IP Detection** - Cloudflare & proxy support
- ✅ **Input Validation** - All user inputs validated

## 🌐 Deployment

### Using Ngrok (Development)

1. Install ngrok: https://ngrok.com/download
2. Start your server: `go run .`
3. Create tunnel: `ngrok http 8080`
4. Update WebSocket URL in code with ngrok URL

### Using Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o omiro .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/omiro .
COPY --from=builder /app/index.html .
EXPOSE 8080
CMD ["./omiro"]
```

Build and run:

```bash
docker build -t omiro .
docker run -p 8080:8080 omiro
```

### Using Docker Compose

```yaml
version: "3.8"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

volumes:
  redis-data:
```

Run with:

```bash
docker-compose up -d
```

### Production Deployment

For production, consider:

- Use **HTTPS/WSS** (required for camera/mic access)
- Set up **Cloudflare** for DDoS protection
- Use **dedicated TURN servers** for better reliability
- Enable **Redis persistence** for session recovery
- Set up **monitoring** (Prometheus/Grafana)
- Configure **log rotation**
- Use **reverse proxy** (Nginx/Caddy)

## 📊 Monitoring

The application tracks:

- Active connections (`stats:active_connections`)
- Queue length
- Match success rate
- Rate limit violations
- Banned IPs

Access stats via Redis:

```bash
redis-cli GET stats:active_connections
redis-cli LLEN matchmaking:queue
```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new features
- Update documentation
- Use meaningful commit messages
- Check for linter errors: `golangci-lint run`

## 🐛 Troubleshooting

### Camera/Microphone Not Working

- Ensure HTTPS is used (required by browsers)
- Check browser permissions
- Verify WebRTC support in browser

### Video Not Connecting

- Check STUN/TURN server configuration
- Verify firewall allows WebRTC traffic
- Check browser console for errors

### WebSocket Connection Fails

- Verify Redis is running
- Check session token is valid
- Ensure rate limits not exceeded

### Redis Connection Error

```bash
# Check Redis is running
redis-cli ping
# Should return: PONG

# Check Redis connection
redis-cli -h localhost -p 6379
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [WebRTC](https://webrtc.org/) - Real-time communication
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [Redis](https://redis.io/) - Data structure store
- [Open Relay](https://www.metered.ca/tools/openrelay/) - Free TURN servers

## 📧 Contact

- **Project Link**: https://github.com/yourusername/omiro
- **Issues**: https://github.com/yourusername/omiro/issues
- **Discussions**: https://github.com/yourusername/omiro/discussions

## ⭐ Star History

If you find this project useful, please consider giving it a star!

---

**Made with ❤️ for the open-source community**
