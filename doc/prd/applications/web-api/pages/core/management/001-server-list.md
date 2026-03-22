# Server List

## Page Description
Entry point page that lists all active log source servers. Each server name is a link to its detail view.

## Access Permissions
Open to all (no authentication).

## Page Structure

### Section 1: Server List
- **Position**: Full page
- **Content**:
  - List of server names as clickable links
- **Interactions**:
  - Click server name → Navigate to Server View page

## Data Display Rules
- **Default Sort**: Alphabetical by server name
- **Empty State**: "No servers configured"

## Page Navigation
- Click server name → `/index/{server-id}` (Server View)
