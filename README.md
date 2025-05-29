# Luma - Terminal API Client

Luma is a minimalist terminal-based API client built with Go. It provides a clean and efficient interface for making HTTP requests and testing APIs directly from your terminal.

## Features

- 🎨 Clean terminal UI with color-coded HTTP methods
- 🔄 Support for GET, POST, PUT, and DELETE requests
- 📝 Request body editor
- 🔑 Header management
- 🔍 Query parameter support
- ⚡ Response display
- ⏱️ Response time tracking
- 📊 Status code visualization


### Controls

#### Navigation
- `Tab` / `Shift+Tab`: Switch between input fields
- `Ctrl+C` or `q`: Exit the application
- `i`: Enter insert mode
- `Esc`: Exit insert mode
- `h`/`l` or `left`/`right`: Switch between request and response sections (in horizontal layout)
- `j`/`k` or `up`/`down`: Switch between sections (in vertical layout)

#### Request Section
- `Tab`: Switch between Body, Headers, and Params tabs
- `Enter`: Submit the request
- `Alt+Backspace`: Delete current header or parameter

#### Headers and Parameters
- `Enter`: Save current header/parameter and move to next
- `Tab`/`Shift+Tab`: Move between headers/parameters
- Maximum of 5 headers/parameters allowed

#### Response Section
- `j`/`k` or `up`/`down`: Scroll through response (in insert mode)

## Interface Layout

The interface is divided into several sections:

1. **Top Bar**
   - HTTP Method selector (GET, POST, PUT, DELETE)
   - URL input field
   - Status code and response time

2. **Request Section**
   - Body tab: Edit request body
   - Headers tab: Manage request headers
   - Params tab: Manage query parameters

3. **Response Section**
   - Displays response body
   - Shows status code and response time
   - Scrollable viewport for large responses

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.