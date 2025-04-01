# MCP Google Search Server

This is an MCP (Model Context Protocol) server that provides a Google Custom Search tool for LLM applications. It allows LLMs to search the web using Google's Custom Search API.

## Features

- Search the web using Google Custom Search API
- Configurable number of results (up to 10)
- Simple and clean result formatting
- Easy integration with LLM applications that support MCP

## Prerequisites

- Go 1.24 or higher
- Google Custom Search API key
- Google Programmable Search Engine ID

## Setup

1. Clone this repository:

   ```
   git clone https://github.com/habuvo/mcp-internet-search.git
   cd mcp-internet-search
   ```

3. Edit the `.env` file and add your Google API credentials (for testing purposes)

2. Update a `env` chapter of your MCP servers `settings.json`:

```
{
  "mcpServers": {
    "google-search": {
      "command": "path_to_executable_binary",
      "args": [],
      "env": {
        "GOOGLE_API_KEY": "your_api_key",
        "GOOGLE_SEARCH_ENGINE_ID": "your_search_engine_id"
      },
      "tools": {
        "google_search": {
          "inputSchema": {
            "type": "object",
            "properties": {
              "query": {
                "type": "string",
                "description": "The search query",
                "required": true
              },
              "num_results": {
                "type": "number",
                "description": "Number of results to return (default: 5, max: 10)",
                "default": 5
              }
            }
          }
        }
      },
      "alwaysAllow": [
        "google_search"
      ],
      "timeout": 15,
      "disabled": true
    }
  }
}
```

   You can get your API key from the [Google Cloud Console](https://console.cloud.google.com/) and create a Programmable Search Engine at [programmablesearchengine.google.com](https://programmablesearchengine.google.com/).

4. Build the server:

   ```

   go build

   ```

## Usage

Run the server:

```

./mcp-internet-search

```

The server will start and listen for MCP requests on stdin/stdout.

### Tool Parameters

The `google_search` tool accepts the following parameters:

- `query` (string, required): The search query
- `num_results` (number, optional): Number of results to return (default: 5, max: 10)

### Example

When integrated with an LLM application that supports MCP, you can use the tool like this:

```

I need information about climate change in Europe.

```

The LLM can then use the `google_search` tool to search for "climate change in Europe" and provide relevant information based on the search results.

## License

MIT
