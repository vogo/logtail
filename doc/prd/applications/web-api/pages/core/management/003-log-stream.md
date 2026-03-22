# Log Stream

## Page Description
WebSocket endpoint for real-time log streaming from a specific server. Supports inline matcher configuration for live filtering.

## Access Permissions
Open to all (no authentication).

## Page Structure

### Section 1: WebSocket Connection
- **Position**: Full page (or embedded in Server View)
- **Content**:
  - Real-time log output streamed via WebSocket
- **Interactions**:
  - Send WebSocket message to configure inline matchers
  - Receive log lines matching current filter

## Feature Descriptions

### Feature 1: Real-time Streaming
- **Trigger**: WebSocket connection to `/tail/{server-id}`
- **Business Logic**: Subscribe to server's log output; apply optional inline matchers
- **Related Models**: Server, Router, Matcher
- **Feedback**: Continuous stream of matching log lines

### Feature 2: Inline Filter Configuration
- **Trigger**: Send matcher configuration via WebSocket message
- **Business Logic**: Apply new matcher conditions to the stream
- **Related Models**: Matcher
- **Feedback**: Stream immediately filtered by new conditions

## Page States
- **Connected**: Actively receiving log lines
- **Disconnected**: WebSocket connection closed
- **No Data**: Connected but no matching log lines
