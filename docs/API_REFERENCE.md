# API Reference Documentation

## Overview

Arch-Manager provides a comprehensive RESTful API for managing users, nodes, system settings, and monitoring. All API endpoints use JSON for request/response data and follow standard HTTP conventions.

## Base URL and Authentication

**Base URL:** `http://your-server:8080/v1/`

**Authentication:** Bearer token authentication
```bash
Authorization: Bearer {token}
```

**Obtaining Token:**
```bash
curl -X POST http://localhost:8080/v1/sign-in \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

**Response:**
```json
{
  "token": "encrypted-jwt-token",
  "expires_at": "2025-08-23T10:30:00Z"
}
```

## Common Response Formats

### Success Response
```json
{
  "success": true,
  "data": { /* response data */ },
  "message": "Operation completed successfully"
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error description",
  "code": 400,
  "details": "Additional error context"
}
```

### HTTP Status Codes
- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Server error

## Authentication Endpoints

### Sign In
**POST** `/v1/sign-in`

**Description:** Authenticate admin user and obtain access token

**Request Body:**
```json
{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2025-08-23T10:30:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/v1/sign-in \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

## User Management Endpoints

### List Users
**GET** `/v1/users`

**Description:** Retrieve all users with their current status and usage

**Query Parameters:**
- `limit` (optional): Number of users per page (default: 100)
- `offset` (optional): Number of users to skip (default: 0)
- `search` (optional): Search by name or ID
- `status` (optional): Filter by enabled/disabled

**Response:**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": 1,
        "identity": "550e8400-e29b-41d4-a716-446655440000",
        "name": "john_doe",
        "quota": 50.0,
        "usage": 15.2,
        "usage_bytes": 16329948160,
        "usage_reset_at": 1692672000000,
        "enabled": true,
        "shadowsocks_password": "randompass123456",
        "shadowsocks_method": "chacha20-ietf-poly1305",
        "created_at": 1692585600000
      }
    ],
    "total": 1,
    "limit": 100,
    "offset": 0
  }
}
```

**Example:**
```bash
curl -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8080/v1/users?limit=10&search=john"
```

### Create User
**POST** `/v1/users`

**Description:** Create a new user account

**Request Body:**
```json
{
  "name": "jane_smith",
  "quota": 100.0,
  "shadowsocks_method": "chacha20-ietf-poly1305"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "identity": "550e8400-e29b-41d4-a716-446655440001",
    "name": "jane_smith",
    "quota": 100.0,
    "usage": 0.0,
    "usage_bytes": 0,
    "usage_reset_at": 1692672000000,
    "enabled": true,
    "shadowsocks_password": "autopass987654321",
    "shadowsocks_method": "chacha20-ietf-poly1305",
    "created_at": 1692672000000
  },
  "message": "User created successfully"
}
```

**Validation Rules:**
- `name`: Required, 1-64 characters, unique
- `quota`: Optional, >= 0 (0 = unlimited)
- `shadowsocks_method`: Optional, valid encryption method

**Example:**
```bash
curl -X POST http://localhost:8080/v1/users \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"jane_smith","quota":100.0}'
```

### Update User
**PUT** `/v1/users/{id}`

**Description:** Update user information

**Path Parameters:**
- `id`: User ID (integer)

**Request Body:**
```json
{
  "name": "jane_smith_updated",
  "quota": 150.0,
  "enabled": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "jane_smith_updated",
    "quota": 150.0,
    "enabled": true
    // ... other fields unchanged
  },
  "message": "User updated successfully"
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/v1/users/2 \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"quota":150.0}'
```

### Partial Update User
**PATCH** `/v1/users/{id}`

**Description:** Partially update user (only specified fields)

**Request Body:**
```json
{
  "enabled": false
}
```

**Response:** Same as PUT endpoint

### Batch Update Users
**PATCH** `/v1/users`

**Description:** Update multiple users at once

**Request Body:**
```json
{
  "ids": [1, 2, 3],
  "updates": {
    "quota": 200.0,
    "enabled": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "updated_count": 3,
    "updated_ids": [1, 2, 3]
  },
  "message": "3 users updated successfully"
}
```

### Delete User
**DELETE** `/v1/users/{id}`

**Description:** Delete a single user

**Response:**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

### Batch Delete Users
**DELETE** `/v1/users`

**Description:** Delete multiple users

**Request Body:**
```json
{
  "ids": [1, 2, 3]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "deleted_count": 3,
    "deleted_ids": [1, 2, 3]
  },
  "message": "3 users deleted successfully"
}
```

## Node Management Endpoints

### List Nodes
**GET** `/v1/nodes`

**Description:** Retrieve all registered nodes with status

**Response:**
```json
{
  "success": true,
  "data": {
    "nodes": [
      {
        "id": 1,
        "host": "192.168.1.100",
        "http_port": 8080,
        "http_token": "secure-token-12345",
        "usage": 125.5,
        "usage_bytes": 134744072192,
        "push_status": "available",
        "pull_status": "available",
        "pushed_at": 1692672000000,
        "pulled_at": 1692671940000
      }
    ],
    "total": 1
  }
}
```

**Node Status Values:**
- `""` (empty): Processing/initial state
- `"available"`: Healthy and operational
- `"dirty"`: Reachable via proxy only
- `"unavailable"`: Unreachable/failed

### Create Node
**POST** `/v1/nodes`

**Description:** Register a new node

**Request Body:**
```json
{
  "host": "192.168.1.101",
  "http_port": 8080,
  "http_token": "secure-node-token-67890"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "host": "192.168.1.101",
    "http_port": 8080,
    "http_token": "secure-node-token-67890",
    "usage": 0.0,
    "usage_bytes": 0,
    "push_status": "",
    "pull_status": "",
    "pushed_at": 0,
    "pulled_at": 0
  },
  "message": "Node created successfully"
}
```

**Validation Rules:**
- `host`: Required, valid IP or hostname
- `http_port`: Required, 1-65536
- `http_token`: Required, authentication token

### Update Node
**PUT** `/v1/nodes/{id}`

**Description:** Update node configuration

**Request Body:**
```json
{
  "host": "192.168.1.102",
  "http_port": 8081,
  "http_token": "updated-token"
}
```

### Batch Update Nodes
**PATCH** `/v1/nodes`

**Description:** Update multiple nodes

**Request Body:**
```json
{
  "ids": [1, 2],
  "updates": {
    "http_port": 8080
  }
}
```

### Delete Node
**DELETE** `/v1/nodes/{id}`

**Description:** Remove a node from the system

**Response:**
```json
{
  "success": true,
  "message": "Node deleted successfully"
}
```

### Get Node Configuration
**GET** `/v1/nodes/{id}/configs`

**Description:** Retrieve generated Xray configuration for specific node

**Response:**
```json
{
  "success": true,
  "data": {
    "config": {
      "log": {
        "loglevel": "info"
      },
      "inbounds": [
        {
          "tag": "direct",
          "port": 8443,
          "protocol": "shadowsocks",
          "settings": {
            "method": "2022-blake3-aes-128-gcm",
            "password": "generated-key-123456"
          }
        }
      ],
      "outbounds": [
        {
          "tag": "out",
          "protocol": "freedom"
        }
      ],
      "metadata": {
        "updated_at": "2025-08-22T10:30:00Z",
        "updated_by": "192.168.1.10"
      }
    }
  }
}
```

## System Management Endpoints

### Get Statistics
**GET** `/v1/stats`

**Description:** Retrieve system-wide statistics

**Response:**
```json
{
  "success": true,
  "data": {
    "total_usage": 1250.5,
    "total_usage_bytes": 1342177280000,
    "total_usage_reset_at": 1692672000000,
    "active_users": 45,
    "total_users": 50,
    "active_nodes": 8,
    "total_nodes": 10,
    "system_info": {
      "uptime": 7200,
      "memory_usage": "156MB",
      "cpu_usage": "12%"
    }
  }
}
```

### Update Statistics
**PATCH** `/v1/stats`

**Description:** Reset or update system statistics

**Request Body:**
```json
{
  "total_usage_reset_at": 1692672000000
}
```

### Get Settings
**GET** `/v1/settings`

**Description:** Retrieve current system settings

**Response:**
```json
{
  "success": true,
  "data": {
    "admin_password": "***hidden***",
    "host": "192.168.1.10",
    "ss_relay_port": 8443,
    "ss_reverse_port": 8444,
    "ss_direct_port": 8445,
    "ss_remote_port": 8446,
    "traffic_ratio": 1.0,
    "reset_policy": "monthly",
    "singet_server": ""
  }
}
```

### Update Settings
**POST** `/v1/settings`

**Description:** Update system configuration

**Request Body:**
```json
{
  "ss_relay_port": 8443,
  "ss_reverse_port": 8444,
  "ss_direct_port": 0,
  "ss_remote_port": 8446,
  "traffic_ratio": 1.5,
  "reset_policy": "monthly",
  "admin_password": "new-secure-password"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Settings updated successfully"
}
```

**Port Configuration:**
- `0`: Disable this connection mode
- `1-65536`: Enable on specified port

### Restart Xray
**POST** `/v1/settings/xray/restart`

**Description:** Restart local Xray service and push configs to nodes

**Response:**
```json
{
  "success": true,
  "message": "Xray restarted successfully"
}
```

**Effects:**
- Regenerates all configurations
- Restarts local Xray-core
- Pushes new configs to all nodes
- Updates routing and load balancing

## Profile Endpoints

### Get User Profile
**GET** `/v1/profile`

**Description:** Get profile information for a specific user

**Query Parameters:**
- `identity`: User UUID (required)

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "john_doe",
      "quota": 50.0,
      "usage": 15.2,
      "enabled": true
    },
    "connections": [
      {
        "name": "Relay Server",
        "url": "ss://chacha20-ietf-poly1305:password@server:8443",
        "qr_code": "data:image/png;base64,..."
      }
    ]
  }
}
```

### Regenerate Profile Links
**POST** `/v1/profile/links/regenerate`

**Description:** Regenerate connection URLs for user

**Request Body:**
```json
{
  "identity": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Information Endpoints

### Get System Information
**GET** `/v1/information`

**Description:** Get system information and license details

**Response:**
```json
{
  "success": true,
  "data": {
    "version": "v25.8.21",
    "xray_version": "Xray v25.8.3",
    "license": {
      "status": "active",
      "max_users": 1024,
      "expires_at": "2026-08-22T00:00:00Z"
    },
    "system": {
      "os": "linux",
      "arch": "amd64",
      "uptime": 7200
    }
  }
}
```

## Import/Export Endpoints

### Import Data
**POST** `/v1/imports`

**Description:** Import users from external sources

**Request Body:**
```json
{
  "source": "subscription_url",
  "url": "https://example.com/subscription",
  "format": "shadowsocks"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "imported_users": 25,
    "skipped_users": 3,
    "errors": []
  },
  "message": "Import completed successfully"
}
```

## Rate Limiting

**Rate Limits:**
- Authentication: 5 requests per minute
- User operations: 100 requests per minute
- Node operations: 50 requests per minute
- Statistics: 200 requests per minute

**Rate Limit Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1692672000
```

## WebSocket API

### Real-time Updates
**WebSocket** `/ws/updates`

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/updates?token=' + token);
```

**Message Format:**
```json
{
  "type": "user_update",
  "data": {
    "user_id": 1,
    "usage": 25.5,
    "enabled": true
  },
  "timestamp": "2025-08-22T10:30:00Z"
}
```

**Message Types:**
- `user_update`: User usage or status change
- `node_update`: Node status change
- `system_update`: System statistics update
- `config_update`: Configuration change notification

## Error Codes Reference

### Validation Errors (422)
```json
{
  "success": false,
  "error": "Validation failed",
  "code": 422,
  "details": {
    "name": ["Name is required"],
    "quota": ["Quota must be greater than 0"]
  }
}
```

### Common Error Codes
- `1001`: User not found
- `1002`: User already exists
- `1003`: Invalid user quota
- `2001`: Node not found
- `2002`: Node connection failed
- `2003`: Invalid node configuration
- `3001`: Invalid settings
- `3002`: Port already in use
- `4001`: Authentication failed
- `4002`: Token expired
- `5001`: Database error
- `5002`: Xray configuration error

## SDK Examples

### JavaScript/Node.js
```javascript
class ArchManagerAPI {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(method, endpoint, data = null) {
    const response = await fetch(`${this.baseURL}/v1${endpoint}`, {
      method,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: data ? JSON.stringify(data) : null
    });

    return response.json();
  }

  async getUsers() {
    return this.request('GET', '/users');
  }

  async createUser(userData) {
    return this.request('POST', '/users', userData);
  }

  async updateUser(id, userData) {
    return this.request('PUT', `/users/${id}`, userData);
  }

  async deleteUser(id) {
    return this.request('DELETE', `/users/${id}`);
  }
}

// Usage
const api = new ArchManagerAPI('http://localhost:8080', 'your-token');
const users = await api.getUsers();
```

### Python
```python
import requests

class ArchManagerAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }

    def request(self, method, endpoint, data=None):
        url = f"{self.base_url}/v1{endpoint}"
        response = requests.request(method, url, headers=self.headers, json=data)
        return response.json()

    def get_users(self):
        return self.request('GET', '/users')

    def create_user(self, user_data):
        return self.request('POST', '/users', user_data)

    def update_user(self, user_id, user_data):
        return self.request('PUT', f'/users/{user_id}', user_data)

    def delete_user(self, user_id):
        return self.request('DELETE', f'/users/{user_id}')

# Usage
api = ArchManagerAPI('http://localhost:8080', 'your-token')
users = api.get_users()
```

### cURL Examples
```bash
# Set token variable
TOKEN="your-auth-token"
BASE_URL="http://localhost:8080/v1"

# List users
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/users"

# Create user
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"testuser","quota":50}' \
  "$BASE_URL/users"

# Update user
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quota":100}' \
  "$BASE_URL/users/1"

# Delete user
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/users/1"

# Get system stats
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/stats"

# Update settings
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ss_relay_port":8443}' \
  "$BASE_URL/settings"
```

This API reference provides complete documentation for all available endpoints, request/response formats, authentication methods, and integration examples.
