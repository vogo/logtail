# Management Console

## Page Description
Runtime configuration management interface. Allows operators to add/remove servers, routers, and transfers without restarting the service.

## Access Permissions
Open to all (no authentication).

## Page Structure

### Section 1: Server Management
- **Position**: Top section
- **Content**:
  - List of configured servers
  - Add server form (file-watch servers only)
  - Delete server button per server
- **Interactions**:
  - Click "Add Server" → Submit server configuration (POST /manage/server/add)
  - Click "Delete" → Remove server (POST /manage/server/delete)

### Section 2: Router Management
- **Position**: Middle section
- **Content**:
  - List of configured routers with matcher and transfer details
  - Add router form
  - Delete router button per router
- **Interactions**:
  - Click "Add Router" → Submit router configuration (POST /manage/router/add)
  - Click "Delete" → Remove router (POST /manage/router/delete)

### Section 3: Transfer Management
- **Position**: Bottom section
- **Content**:
  - List of configured transfers with type and URL
  - Add transfer form
  - Delete transfer button per transfer
- **Interactions**:
  - Click "Add Transfer" → Submit transfer configuration (POST /manage/transfer/add)
  - Click "Delete" → Remove transfer (POST /manage/transfer/delete)

### Section 4: Statistics
- **Position**: Bottom of page or separate tab
- **Content**:
  - Per-router pipeline statistics: drop count, buffer size, blocking mode
- **Interactions**:
  - Refresh statistics (GET /manage/stats)

## Feature Descriptions

### Feature 1: Add Server
- **Trigger**: Submit add server form
- **Business Logic**: Only file-watch servers can be added via API. Configuration is persisted to disk.
- **Related Models**: Server, ServerConfig
- **Feedback**: Success message or error

### Feature 2: Add Router
- **Trigger**: Submit add router form
- **Business Logic**: Create router with specified matchers and transfer references. Configuration persisted.
- **Related Models**: Router, RouterConfig
- **Feedback**: Success message or error

### Feature 3: Add Transfer
- **Trigger**: Submit add transfer form
- **Business Logic**: Create transfer of specified type. Configuration persisted.
- **Related Models**: Transfer, TransferConfig
- **Feedback**: Success message or error

### Feature 4: Delete Component
- **Trigger**: Click delete button
- **Business Logic**: Stop the component, remove from pipeline. Transfers in use by routers cannot be deleted.
- **Related Models**: Server, Router, Transfer
- **Feedback**: Success or "transfer in use" error

## Business Rules
- Only file-watch servers can be added via API (no command-based servers)
- Transfers referenced by active routers cannot be deleted
- All changes are persisted to configuration file

## Data Display Rules
- **Default Sort**: Alphabetical by name
- **Empty State**: "No [servers/routers/transfers] configured"
- **Error State**: Display error message from API response
