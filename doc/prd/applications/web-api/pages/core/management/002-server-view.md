# Server View

## Page Description
Detail page for a single server showing embedded monitoring UI.

## Access Permissions
Open to all (no authentication).

## Page Structure

### Section 1: Server Information
- **Position**: Top of page
- **Content**:
  - Server name and identifier
- **Interactions**:
  - Navigate back to Server List

### Section 2: Log Display
- **Position**: Main content area
- **Content**:
  - Embedded log output view
- **Interactions**:
  - Auto-scrolling log display

## Page Navigation
- Back → `/` (Server List)
- WebSocket connection → `/tail/{server-id}` (Log Stream)
