# Telex Agora Token Server

A lightweight Go-based token server for generating Agora RTC (Real-Time Communication) and RTM (Real-Time Messaging) tokens. This service provides secure token generation endpoints for Agora-powered voice, video, and messaging applications.

## Features

- 🔐 Secure token generation using Agora App ID and Certificate
- 🎥 **RTC Token Generation** - For voice and video channels
- 💬 **RTM Token Generation** - For real-time messaging
- 🔄 **Combined Token Generation** - Get both RTC and RTM tokens in one request
- ⚡ Built with Gin framework for high performance
- 🔧 Configurable token expiration times
- 📦 Support for both UID and user account-based tokens

## Prerequisites

- Go 1.23.0 or higher
- Agora App ID and App Certificate ([Get them here](https://console.agora.io/))

## Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/telexorg/telex_agora.git
   cd telex_agora
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   
   Copy the example environment file and add your Agora credentials:
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and add your credentials:
   ```env
   APP_ID=your_agora_app_id_here
   APP_CERTIFICATE=your_agora_app_certificate_here
   ```

## Running the Server

Start the server on port 8080:

```bash
go run main.go
```

The server will be available at `http://localhost:8080`

## API Endpoints

### Health Check

```
GET /ping
```

Returns a simple pong response to verify the server is running.

**Response:**
```json
{
  "message": "pong"
}
```

---

### Generate RTC Token

```
GET /rtc/:channelName/:role/:tokenType/:uid?expiry=3600
```

Generate a token for Agora Real-Time Communication (voice/video).

**Parameters:**
- `channelName` (path) - The name of the channel
- `role` (path) - User role: `publisher` or `subscriber`
- `tokenType` (path) - Token type: `uid` or `userAccount`
- `uid` (path) - User ID (numeric if tokenType is `uid`, string if `userAccount`)
- `expiry` (query, optional) - Token expiration time in seconds (default: 3600)

**Examples:**

Generate token with UID:
```bash
curl http://localhost:8080/rtc/myChannel/publisher/uid/12345?expiry=7200
```

Generate token with user account:
```bash
curl http://localhost:8080/rtc/myChannel/publisher/userAccount/john_doe?expiry=3600
```

**Response:**
```json
{
  "rtcToken": "006..."
}
```

---

### Generate RTM Token

```
GET /rtm/:uid?expiry=3600
```

Generate a token for Agora Real-Time Messaging.

**Parameters:**
- `uid` (path) - User ID
- `expiry` (query, optional) - Token expiration time in seconds (default: 3600)

**Example:**
```bash
curl http://localhost:8080/rtm/john_doe?expiry=3600
```

**Response:**
```json
{
  "rtmToken": "006..."
}
```

---

### Generate Both RTC and RTM Tokens

```
GET /rte/:channelName/:role/:tokenType/:uid?expiry=3600
```

Generate both RTC and RTM tokens in a single request.

**Parameters:**
- Same as RTC token generation

**Example:**
```bash
curl http://localhost:8080/rte/myChannel/publisher/uid/12345?expiry=3600
```

**Response:**
```json
{
  "rtcToken": "006...",
  "rtmToken": "006..."
}
```

## Token Types

### UID-based Tokens
Use numeric user IDs. Best for scenarios where you manage users with numeric identifiers.

```
/rtc/gameRoom/publisher/uid/12345
```

### UserAccount-based Tokens
Use string-based user accounts. Ideal when you have string-based user identifiers.

```
/rtc/gameRoom/publisher/userAccount/john_doe
```

## Role Types

- **publisher** - Can publish and subscribe to streams (send and receive audio/video)
- **subscriber** - Can only subscribe to streams (receive only)

## Error Handling

The API returns appropriate HTTP status codes and error messages:

**400 Bad Request** - Invalid parameters or token generation failure
```json
{
  "status": 400,
  "error": "Error generating RTC token: ..."
}
```

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [Agora Go Token Builder](https://github.com/AgoraIO-Community/go-tokenbuilder) - Official Agora token generation library
- [godotenv](https://github.com/joho/godotenv) - Environment variable management

## Security Considerations

- ⚠️ Never commit your `.env` file or expose your App Certificate
- 🔒 Keep your App ID and App Certificate secure
- 🌐 Use HTTPS in production environments
- 🛡️ Implement rate limiting for production deployments
- 🔐 Consider adding authentication middleware to protect token generation endpoints

## Development

To modify the server port, edit the last line in `main.go`:

```go
api.Run(":8080") // Change to your desired port
```

## License

This project is licensed under the MIT License.

## Resources

- [Agora Documentation](https://docs.agora.io/)