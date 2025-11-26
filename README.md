# Telex Agora Huddle Backend

A backend service for building real-time audio huddles (Slack Huddles-like) using Agora. Provides token generation and huddle management for your frontend application.

## For Frontend Developers

This service provides everything you need to build a real-time audio huddle application. Choose the integration approach that fits your needs.

---

## 🚀 Quick Start for Frontend

### Base URL
```
http://localhost:8080
```

### Two Ways to Build Your App

#### Option 1: Automatic (Recommended for Quick Start)
Just request a token - the backend handles everything else automatically.

#### Option 2: Explicit Control
Use dedicated endpoints when you need full control over huddle lifecycle.

---

## 📖 Frontend Integration Guide

### Approach 1: Automatic (Simplest - Recommended)

**Perfect for**: Simple huddle apps, quick prototypes, MVPs

```javascript
import AgoraRTC from 'agora-rtc-sdk-ng';

async function joinHuddle(roomName, userId) {
  // Step 1: Get token (huddle created automatically!)
  const response = await fetch(
    `http://localhost:8080/rtc/${roomName}/publisher/userAccount/${userId}`
  );

  const { rtcToken, appId, channelName, huddleId } = await response.json();

  // Step 2: Join Agora channel
  const client = AgoraRTC.createClient({ mode: 'rtc', codec: 'vp8' });
  await client.join(appId, channelName, rtcToken, userId);

  // Step 3: Publish audio
  const audioTrack = await AgoraRTC.createMicrophoneAudioTrack();
  await client.publish([audioTrack]);

  console.log('Joined huddle:', huddleId);
  return { client, audioTrack, huddleId };
}

// Usage
await joinHuddle('team-standup', 'alice');
```

---

### Approach 2: Explicit Control (Advanced)

**Perfect for**: Complex apps, custom UI flows, advanced tracking

```javascript
import AgoraRTC from 'agora-rtc-sdk-ng';

async function createAndJoinHuddle(userId) {
  // Step 1: Create huddle
  const createRes = await fetch('http://localhost:8080/huddle/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ created_by: userId })
  });

  const { huddle_id, channel_name, app_id } = await createRes.json();

  // Step 2: Join huddle (tracking)
  await fetch('http://localhost:8080/huddle/join', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ huddle_id, user_id: userId })
  });

  // Step 3: Get token
  const tokenRes = await fetch(
    `http://localhost:8080/rtc/${channel_name}/publisher/userAccount/${userId}`
  );

  const { rtcToken } = await tokenRes.json();

  // Step 4: Join Agora
  const client = AgoraRTC.createClient({ mode: 'rtc', codec: 'vp8' });
  await client.join(app_id, channel_name, rtcToken, userId);

  // Step 5: Publish audio
  const audioTrack = await AgoraRTC.createMicrophoneAudioTrack();
  await client.publish([audioTrack]);

  return { client, audioTrack, huddle_id, channel_name };
}

async function leaveHuddle(huddleId, userId, client, audioTrack) {
  // Clean up Agora
  await audioTrack.close();
  await client.leave();

  // Leave huddle on backend
  await fetch('http://localhost:8080/huddle/leave', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ huddle_id: huddleId, user_id: userId })
  });
}

// Usage
const { client, audioTrack, huddle_id } = await createAndJoinHuddle('alice');
// ... later
await leaveHuddle(huddle_id, 'alice', client, audioTrack);
```

---

### Display Active Huddles

```javascript
async function getActiveHuddles() {
  const response = await fetch('http://localhost:8080/huddles');
  const { huddles } = await response.json();
  return huddles;
}

// Usage
const huddles = await getActiveHuddles();
huddles.forEach(huddle => {
  console.log(`${huddle.channel_name}: ${huddle.participant_count} users`);
  console.log(`Participants: ${huddle.participants.join(', ')}`);
});
```

---

## 📡 API Endpoints Reference

### Token Endpoints (Automatic Huddle Management)

#### Get RTC Token
```
GET /rtc/:channelName/:role/:tokenType/:uid?expiry=3600
```

**What it does:**
- Generates Agora RTC token
- Creates huddle automatically if it doesn't exist
- Tracks user as participant

**Parameters:**
- `channelName` - Your room/channel name (e.g., "team-standup")
- `role` - `publisher` (can send/receive) or `subscriber` (receive only)
- `tokenType` - `userAccount` (for string IDs) or `uid` (for numeric IDs)
- `uid` - User identifier (e.g., "alice" or "12345")
- `expiry` - Optional, token expiration in seconds (default: 3600)

**Response:**
```json
{
  "rtcToken": "006abc123...",
  "channelName": "team-standup",
  "huddleId": "uuid-here",
  "appId": "your_app_id"
}
```

**Example:**
```javascript
const response = await fetch(
  'http://localhost:8080/rtc/my-room/publisher/userAccount/alice'
);
```

---

#### Get Both RTC + RTM Tokens
```
GET /rte/:channelName/:role/:tokenType/:uid?expiry=3600
```

**Same as above but returns both RTC and RTM tokens:**
```json
{
  "rtcToken": "006abc...",
  "rtmToken": "006def...",
  "channelName": "team-standup",
  "huddleId": "uuid-here",
  "appId": "your_app_id"
}
```

---

### Explicit Huddle Management Endpoints

#### Create Huddle
```
POST /huddle/create
```

**Request:**
```json
{
  "created_by": "alice"
}
```

**Response:**
```json
{
  "huddle_id": "uuid-here",
  "channel_name": "huddle_uuid",
  "created_by": "alice",
  "created_at": "2025-11-25T17:23:01Z",
  "app_id": "your_app_id"
}
```

---

#### Join Huddle
```
POST /huddle/join
```

**Request:**
```json
{
  "huddle_id": "uuid-here",
  "user_id": "bob"
}
```

**Response:**
```json
{
  "message": "Successfully joined huddle"
}
```

---

#### Leave Huddle
```
POST /huddle/leave
```

**Request:**
```json
{
  "huddle_id": "uuid-here",
  "user_id": "bob"
}
```

**Response:**
```json
{
  "message": "Successfully left huddle"
}
```

---

#### List Active Huddles
```
GET /huddles
GET /huddle/list (alternative)
```

**Response:**
```json
{
  "huddles": [
    {
      "huddle_id": "uuid-here",
      "channel_name": "my-room",
      "created_by": "alice",
      "created_at": "2025-11-25T17:23:01Z",
      "participant_count": 2,
      "participants": ["alice", "bob"]
    }
  ]
}
```

---

#### End Huddle
```
DELETE /huddles/:channelName
POST /huddle/end (with body: {"huddle_id": "uuid"})
```

**Response:**
```json
{
  "message": "Huddle ended successfully"
}
```

---

## 🎨 Complete React Example

```jsx
import { useState, useEffect } from 'react';
import AgoraRTC from 'agora-rtc-sdk-ng';

const BACKEND_URL = 'http://localhost:8080';

function HuddleApp() {
  const [huddles, setHuddles] = useState([]);
  const [client, setClient] = useState(null);
  const [audioTrack, setAudioTrack] = useState(null);
  const [currentHuddleId, setCurrentHuddleId] = useState(null);
  const [userId] = useState('user_' + Date.now());

  // Fetch active huddles
  useEffect(() => {
    const fetchHuddles = async () => {
      const res = await fetch(`${BACKEND_URL}/huddles`);
      const data = await res.json();
      setHuddles(data.huddles || []);
    };

    fetchHuddles();
    const interval = setInterval(fetchHuddles, 5000);
    return () => clearInterval(interval);
  }, []);

  // Join huddle (automatic approach)
  const joinHuddle = async (roomName) => {
    try {
      // Get token (huddle created automatically)
      const res = await fetch(
        `${BACKEND_URL}/rtc/${roomName}/publisher/userAccount/${userId}`
      );
      const { rtcToken, appId, channelName, huddleId } = await res.json();

      // Join Agora
      const agoraClient = AgoraRTC.createClient({ mode: 'rtc', codec: 'vp8' });
      await agoraClient.join(appId, channelName, rtcToken, userId);

      // Publish audio
      const audio = await AgoraRTC.createMicrophoneAudioTrack();
      await agoraClient.publish([audio]);

      setClient(agoraClient);
      setAudioTrack(audio);
      setCurrentHuddleId(huddleId);

      console.log('Joined huddle:', huddleId);
    } catch (error) {
      console.error('Failed to join huddle:', error);
    }
  };

  // Leave huddle
  const leaveHuddle = async () => {
    if (audioTrack) await audioTrack.close();
    if (client) await client.leave();

    setClient(null);
    setAudioTrack(null);
    setCurrentHuddleId(null);
  };

  return (
    <div>
      <h1>Audio Huddles</h1>

      {/* Join huddle */}
      <div>
        <input id="roomInput" placeholder="Room name" />
        <button onClick={() => {
          const room = document.getElementById('roomInput').value;
          joinHuddle(room);
        }}>
          Join Huddle
        </button>
        {currentHuddleId && (
          <button onClick={leaveHuddle}>Leave Huddle</button>
        )}
      </div>

      {/* Active huddles list */}
      <h2>Active Huddles</h2>
      {huddles.map(huddle => (
        <div key={huddle.huddle_id}>
          <strong>{huddle.channel_name}</strong>
          <p>{huddle.participant_count} participants: {huddle.participants.join(', ')}</p>
          <button onClick={() => joinHuddle(huddle.channel_name)}>
            Join
          </button>
        </div>
      ))}
    </div>
  );
}

export default HuddleApp;
```

---

## 📊 Comparison: Which Approach to Use?

| Feature | Automatic | Explicit |
|---------|-----------|----------|
| **API calls** | 1 | 2-3 |
| **Setup time** | Fastest | More setup |
| **Control** | Less | Full control |
| **Best for** | Simple apps, MVPs | Complex workflows |
| **Tracking** | Automatic | Manual |

**Recommendation**: Start with the **Automatic approach** for faster development, then migrate to Explicit if you need more control.

---

## 🔧 Frontend Setup

### Install Agora SDK

```bash
npm install agora-rtc-sdk-ng
# or
yarn add agora-rtc-sdk-ng
```

### CORS

The backend has CORS enabled for all origins during development. For production, configure appropriate origins.

---

## 🧪 Test the Backend

```bash
# Check if backend is running
curl http://localhost:8080/ping

# Test token generation
curl http://localhost:8080/rtc/test-room/publisher/userAccount/testuser

# List active huddles
curl http://localhost:8080/huddles
```

---

## 📚 Additional Resources

- [Full API Documentation](API_REFERENCE.md) - Complete endpoint reference
- [Agora Web SDK Docs](https://docs.agora.io/en/voice-calling/get-started/get-started-sdk) - Agora integration guide
- [Agora Console](https://console.agora.io/) - Get your Agora credentials

---

## ⚠️ Important Notes

- **Tokens expire after 1 hour by default** - Request new tokens when needed
- **In-memory storage** - Huddles are lost on server restart (perfect for development)
- **User IDs** - Use consistent string identifiers (e.g., email, username, UUID)
- **Channel names** - Use URL-safe strings (letters, numbers, hyphens, underscores)

---

## 🚨 Common Issues

### "Invalid token" error
- Token may have expired (default: 1 hour)
- Request a new token from the backend

### Can't hear other users
- Ensure you're using `publisher` role (not `subscriber`)
- Check microphone permissions in browser
- Verify other users have published audio tracks

### CORS errors
- Ensure backend is running on `http://localhost:8080`
- Check that CORS is properly configured for your frontend origin

---

## Backend Setup (for reference)

The backend requires:

```env
APP_ID=your_agora_app_id
APP_CERTIFICATE=your_agora_app_certificate
PORT=8080  # optional
```

Start the backend:
```bash
go run main.go
```

---

## 🎉 You're Ready!

Choose your approach and start building:

1. **Quick prototype?** → Use Automatic approach
2. **Complex app?** → Use Explicit approach
3. **Not sure?** → Start with Automatic, migrate later if needed

Happy coding! 🎙️
