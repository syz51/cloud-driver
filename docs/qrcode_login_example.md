# 115driver QR Code Login API

This document demonstrates how to use the new QR code login functionality for 115 cloud drive integration.

## API Endpoints

All endpoints are prefixed with `/api/v1/115/qrcode/`

### 1. Start QR Code Session

**POST** `/api/v1/115/qrcode/start`

Initiates a new QR code login session.

**Request Body:**

```json
{}
```

**Response:**

```json
{
  "uid": "12345678",
  "sign": "abcdef123456",
  "time": 1703123456789,
  "qrcode_content": "https://qrcode.115.com/api/1.0/web/1.0/token/..."
}
```

### 2. Get QR Code Image

**POST** `/api/v1/115/qrcode/image`

Generates and returns the QR code image as PNG.

**Request Body:**

```json
{
  "uid": "12345678"
}
```

**Response:**

- Content-Type: `image/png`
- Returns binary PNG image data

### 3. Check QR Code Status

**POST** `/api/v1/115/qrcode/status`

Checks the current scan status of the QR code.

**Request Body:**

```json
{
  "uid": "12345678",
  "sign": "abcdef123456",
  "time": 1703123456789
}
```

**Response:**

```json
{
  "status": 0,
  "message": "Waiting for scan",
  "version": "1.0"
}
```

**Status Values:**

- `0`: Waiting for scan
- `1`: QR code scanned, waiting for confirmation
- `2`: Login confirmed, ready to complete
- `-1`: QR code expired
- `-2`: Login canceled

### 4. Complete Login

**POST** `/api/v1/115/qrcode/login`

Completes the QR code login and returns credentials.

**Request Body:**

```json
{
  "uid": "12345678",
  "sign": "abcdef123456",
  "time": 1703123456789,
  "app": "web"
}
```

**Optional App Values:**

- `web` (default)
- `android`
- `ios`
- `tv`
- `alipaymini`
- `wechatmini`
- `qandroid`

**Response (Success):**

```json
{
  "credentials": {
    "uid": "987654321",
    "cid": "xyz789abc",
    "seid": "session123",
    "kid": "key456"
  },
  "success": true,
  "message": "Login successful"
}
```

**Response (Failure):**

```json
{
  "credentials": {
    "uid": "",
    "cid": "",
    "seid": "",
    "kid": ""
  },
  "success": false,
  "message": "QR code not ready for login. Current status: waiting for scan"
}
```

## Usage Flow

1. **Start Session**: Call `/qrcode/start` to get QR code session data
2. **Display QR Code**: Use the `uid` to call `/qrcode/image` and display the QR code to the user
3. **Poll Status**: Regularly call `/qrcode/status` to check if the user has scanned and confirmed
4. **Complete Login**: When status is `2` (allowed), call `/qrcode/login` to get credentials
5. **Use Credentials**: Use the returned credentials for subsequent 115drive API calls

## Example JavaScript Flow

```javascript
async function qrCodeLogin() {
  // 1. Start QR session
  const startResponse = await fetch("/api/v1/115/qrcode/start", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  const session = await startResponse.json();

  // 2. Display QR code image
  const qrImg = document.getElementById("qr-code");
  qrImg.src = `/api/v1/115/qrcode/image`;
  // Include UID in request body for the image endpoint

  // 3. Poll for status
  const pollStatus = async () => {
    const statusResponse = await fetch("/api/v1/115/qrcode/status", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        uid: session.uid,
        sign: session.sign,
        time: session.time,
      }),
    });
    const status = await statusResponse.json();

    if (status.status === 2) {
      // 4. Complete login
      const loginResponse = await fetch("/api/v1/115/qrcode/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          uid: session.uid,
          sign: session.sign,
          time: session.time,
          app: "web",
        }),
      });
      const result = await loginResponse.json();

      if (result.success) {
        console.log("Login successful!", result.credentials);
        // Store credentials for future API calls
        localStorage.setItem(
          "115credentials",
          JSON.stringify(result.credentials)
        );
      }
    } else if (status.status < 0) {
      console.log("QR code expired or canceled:", status.message);
    } else {
      // Continue polling
      setTimeout(pollStatus, 2000);
    }
  };

  pollStatus();
}
```

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200`: Success
- `400`: Bad request (validation errors, QR code not ready)
- `500`: Internal server error

Error responses follow this format:

```json
{
  "message": "Error description"
}
```
