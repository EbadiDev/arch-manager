# Admin Panel Documentation

## Overview

The Arch-Manager admin panel provides a comprehensive web interface for managing users, nodes, system settings, and monitoring performance. The panel is built with modern web technologies and provides real-time updates.

## Access and Authentication

### Login Process

**URL:** `http://your-server:8080`

**Default Credentials:**
- Username: `admin`
- Password: `password`

**Authentication Flow:**
1. User enters credentials on login page
2. System validates against database settings (`AdminPassword`)
3. Server generates encrypted token using Enigma (Ed25519)
4. Token stored in browser session/localStorage
5. All subsequent API calls include Bearer token

### Security Features

- **Session Management**: Automatic logout on token expiration
- **CSRF Protection**: Built-in cross-site request forgery protection
- **Password Security**: Admin password stored in database, changeable via settings
- **Token Encryption**: JWT-style tokens with Ed25519 encryption

## Panel Structure

### Main Navigation

```
┌─────────────────────────────────────────────────────────────┐
│  🌐 Arch-Manager                                  admin ▼   │
├─────────────────────────────────────────────────────────────┤
│  👥 Users  │  🖥️ Nodes  │  📊 System  │  📈 Insights      │
└─────────────────────────────────────────────────────────────┘
```

**Available Sections:**
- **Users** (`/users`): User account management
- **Nodes** (`/nodes`): Server node management  
- **System** (`/system`): System settings and configuration
- **Insights** (`/insights`): Analytics and statistics

## Users Management (`/users`)

### User Interface

**Main Features:**
- **User List**: Paginated table with all users
- **Search/Filter**: Find users by name, ID, or status
- **Bulk Operations**: Select multiple users for batch actions
- **Real-time Stats**: Live usage and quota information

### User Operations

#### 1. Add New User

**Interface:** Modal dialog with form fields

**Required Fields:**
```
Name: [User display name]
Quota: [Monthly limit in GB] (0 = unlimited)
Method: [Encryption method] (dropdown)
```

**API Call:**
```javascript
POST /v1/users
{
  "name": "john_doe",
  "quota": 50.0,
  "shadowsocks_method": "chacha20-ietf-poly1305"
}
```

**Auto-generated Fields:**
- ID: Sequential number (1, 2, 3...)
- Identity: UUID for Xray configurations
- Password: 16-character random string
- Created timestamp

#### 2. Edit User

**Available Edits:**
- Name change
- Quota adjustment (GB)
- Enable/disable status
- Encryption method

**Bulk Edit:** Select multiple users and modify quotas simultaneously

**API Call:**
```javascript
PUT /v1/users/{id}
{
  "name": "john_doe_updated",
  "quota": 100.0,
  "enabled": true
}
```

#### 3. User Profile Links

Each user has access links generated automatically:

**Client Configuration URLs:**
```
Shadowsocks: ss://method:password@server:port
QR Code: Automatically generated for mobile apps
JSON Config: Complete configuration file
```

**Link Generation:**
```javascript
POST /v1/profile/links/regenerate
{
  "user_id": 123
}
```

#### 4. Delete Users

**Single Delete:** Click delete button on user row
**Batch Delete:** Select multiple users and delete all

**API Call:**
```javascript
DELETE /v1/users/{id}
// or
DELETE /v1/users
{
  "ids": [1, 2, 3]
}
```

### User Status Indicators

**Status Colors:**
- 🟢 **Green**: Active user, within quota
- 🟡 **Yellow**: Active user, approaching quota (80%+)
- 🔴 **Red**: Disabled user (over quota or manually disabled)
- ⚫ **Gray**: Disabled user (manually disabled)

**Usage Display:**
```
[████████░░] 45.2 GB / 50.0 GB (90.4%)
```

## Nodes Management (`/nodes`)

### Node Interface

**Main Features:**
- **Node List**: All registered nodes with status
- **Health Monitoring**: Real-time connection status
- **Configuration Management**: Push configs to nodes
- **Statistics**: Traffic usage per node

### Node Operations

#### 1. Add New Node

**Interface:** Modal with connection details

**Required Fields:**
```
Host: [IP address or hostname]
HTTP Port: [API port number] (default: 8080)
HTTP Token: [Authentication token]
```

**API Call:**
```javascript
POST /v1/nodes
{
  "host": "192.168.1.100",
  "http_port": 8080,
  "http_token": "secure-random-token"
}
```

#### 2. Node Status Monitoring

**Push Status** (Configuration sync):
- 🟢 **Available**: Config successfully pushed
- 🟡 **Dirty**: Config pushed via proxy
- 🔴 **Unavailable**: Cannot push config
- ⚪ **Processing**: Config update in progress

**Pull Status** (Health checks):
- 🟢 **Available**: Node responding to health checks
- 🔴 **Unavailable**: Node not responding (>1 minute)
- ⚪ **Processing**: Initial connection

#### 3. Node Configuration

**View Config:** See generated Xray configuration for node
```javascript
GET /v1/nodes/{id}/configs
```

**Force Sync:** Manually push configuration to node
```javascript
POST /v1/settings/xray/restart
```

#### 4. Remove Node

**API Call:**
```javascript
DELETE /v1/nodes/{id}
```

**Effects:**
- Node removed from database
- Traffic routing updated
- Load balancer reconfigured

## System Settings (`/system`)

### System Overview

**Server Information:**
- CPU usage and memory consumption
- Network traffic statistics
- Active connections count
- System uptime and version

**Service Status:**
- Arch-Manager service status
- Xray-core status and version
- Database health and backup status

### Configuration Settings

#### 1. Network Configuration

**Connection Ports:**
```
Relay Port: [Port for relay mode] (0 = disabled)
Reverse Port: [Port for reverse mode] (0 = disabled)
Direct Port: [Port for direct mode] (0 = disabled)
Remote Port: [Port for remote node access] (0 = disabled)
```

**API Call:**
```javascript
POST /v1/settings
{
  "ss_relay_port": 8443,
  "ss_reverse_port": 8444,
  "ss_direct_port": 8445,
  "ss_remote_port": 8446
}
```

#### 2. Traffic Settings

**Traffic Accounting:**
```
Traffic Ratio: [Multiplier for usage calculation] (default: 1.0)
Reset Policy: [monthly/never] (when to reset usage)
```

#### 3. Proxy Settings

**Singet Server:** Proxy server for node communication when direct connection fails
```
Singet Server: [proxy-server:port] (optional)
```

#### 4. Security Settings

**Admin Password:** Change admin panel password
```javascript
POST /v1/settings
{
  "admin_password": "new-secure-password"
}
```

### System Operations

#### 1. Xray Management

**Restart Xray:** Restart local Xray-core instance
```javascript
POST /v1/settings/xray/restart
```

**Effects:**
- Regenerates all configurations
- Restarts local Xray service
- Pushes new configs to all nodes
- Updates routing and load balancing

#### 2. Database Operations

**View Stats:**
```javascript
GET /v1/stats
{
  "total_usage": 1250.5,
  "total_usage_bytes": 1342177280000,
  "total_usage_reset_at": 1692672000000
}
```

**Reset Statistics:**
```javascript
PATCH /v1/stats
{
  "total_usage_reset_at": 1692672000000
}
```

## Insights and Analytics (`/insights`)

### Usage Statistics

**System Overview:**
- Total bandwidth consumed (all users/nodes)
- Active user count vs total users
- Node health and availability
- Peak usage times and trends

**User Analytics:**
- Top consumers by bandwidth
- Quota utilization rates
- Monthly usage patterns
- Geographic distribution (if available)

**Node Analytics:**
- Per-node traffic distribution
- Load balancing effectiveness
- Connection success rates
- Performance metrics

### Charts and Graphs

**Traffic Over Time:**
- Real-time bandwidth usage
- Historical trends (daily/weekly/monthly)
- User vs node traffic breakdown

**User Statistics:**
- Usage distribution histogram
- Quota vs actual usage
- Active vs inactive users

## Real-time Updates

### WebSocket Integration

The admin panel uses WebSocket connections for real-time updates:

**Auto-refresh Components:**
- User usage statistics (every 30 seconds)
- Node status indicators (every 10 seconds)
- System statistics (every 60 seconds)
- Connection counts (every 15 seconds)

### Push Notifications

**System Alerts:**
- Node disconnection warnings
- User quota exceeded notifications
- System error alerts
- Maintenance reminders

## Mobile Responsiveness

### Responsive Design

**Desktop View (>1200px):**
- Full navigation sidebar
- Multi-column layouts
- Detailed statistics panels
- Advanced filtering options

**Tablet View (768px-1200px):**
- Collapsible sidebar
- Responsive tables
- Touch-optimized controls

**Mobile View (<768px):**
- Bottom navigation
- Single-column layout
- Swipe gestures
- Simplified interface

### Touch Optimizations

- Large touch targets (44px minimum)
- Swipe gestures for table navigation
- Pull-to-refresh functionality
- Optimized form inputs

## API Integration

### Frontend-Backend Communication

**API Base URL:** `/v1/`

**Authentication Header:**
```javascript
headers: {
  'Authorization': `Bearer ${localStorage.getItem('token')}`,
  'Content-Type': 'application/json'
}
```

**Common API Patterns:**

**GET Requests** (Data retrieval):
```javascript
const users = await fetch('/v1/users').then(r => r.json());
const nodes = await fetch('/v1/nodes').then(r => r.json());
const stats = await fetch('/v1/stats').then(r => r.json());
```

**POST Requests** (Create operations):
```javascript
const newUser = await fetch('/v1/users', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'test', quota: 50 })
});
```

**PUT/PATCH Requests** (Update operations):
```javascript
const updatedUser = await fetch(`/v1/users/${id}`, {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'updated', quota: 100 })
});
```

### Error Handling

**HTTP Status Codes:**
- `200`: Success
- `400`: Bad Request (validation errors)
- `401`: Unauthorized (invalid token)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found (resource doesn't exist)
- `500`: Internal Server Error

**Error Response Format:**
```json
{
  "error": "User not found",
  "code": 404,
  "details": "User with ID 123 does not exist"
}
```

## Browser Compatibility

### Supported Browsers

**Desktop:**
- Chrome 80+ ✅
- Firefox 75+ ✅
- Safari 13+ ✅
- Edge 80+ ✅

**Mobile:**
- Chrome Mobile 80+ ✅
- Safari Mobile 13+ ✅
- Firefox Mobile 75+ ✅

### Required Features

- ES6+ JavaScript support
- Fetch API
- WebSocket support
- CSS Grid and Flexbox
- Local Storage

## Performance Optimization

### Loading Performance

- **Code Splitting**: Lazy load panel sections
- **Asset Optimization**: Minified CSS/JS
- **Caching**: Aggressive browser caching
- **CDN**: Bootstrap and jQuery from CDN

### Runtime Performance

- **Virtual Scrolling**: Large user/node lists
- **Debounced Search**: Optimized filtering
- **Background Updates**: Non-blocking API calls
- **Memory Management**: Cleanup unused components

This admin panel provides a comprehensive, user-friendly interface for managing all aspects of the Arch-Manager system while maintaining high performance and accessibility.
