# Razorpay MCP Server

The Razorpay MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with Razorpay APIs, enabling advanced payment processing capabilities for developers and AI tools.

## Available Tools

Currently, the Razorpay MCP Server provides the following tools:

| Tool                  | Description                           |
|-----------------------|---------------------------------------|
| `payment.fetch`       | Fetch payment details                 |
| `payment_link.create` | Creates a new payment link            |
| `payment_link.fetch`  | Fetch details of a payment link       |
| `order.create`        | Creates an order                      |
| `order.fetch`         | Fetch order details                   |


## Use Cases 
- Workflow Automation: Automate your day to day workflow using Razorpay MCP Server.
- Agentic Applications: Building AI powered tools that interact with Razorpay's payment ecosystem using this Razorpay MCP server.

## Setup

### Prerequisites
- Docker
- Golang (Go)
- Git

To run the Razorpay MCP server, use one of the following methods:

### Using Docker (Recommended)

You need to clone the Github repo and build the image for Razorpay MCP Server using `docker`. Do make sure `docker` is installed and running in your system. 

```bash
# Run the server
git clone https://github.com/razorpay/razorpay-mcp-server.git
cd razorpay-mcp-server
docker build -t razorpay-mcp-server:latest .
```

Post this razorpay-mcp-server:latest docker image would be ready in your system.

### Build from source

```bash
# Clone the repository
git clone https://github.com/razorpay/razorpay-mcp-server.git
cd razorpay-mcp-server

# Build the binary
go build -o razorpay-mcp-server ./cmd/razorpay-mcp-server
```

Binary `razorpay-mcp-server` would be present in your system post this.

## Usage with Claude Desktop

Add the following to your `claude_desktop_config.json`:

```json
{
    "mcpServers": {
        "razorpay-mcp-server": {
            "command": "docker",
            "args": [
                "run",
                "--rm",
                "-i",
                "-e",
                "RAZORPAY_KEY_ID",
                "-e",
                "RAZORPAY_KEY_SECRET",
                "razorpay-mcp-server:latest"
            ],
            "env": {
                "RAZORPAY_KEY_ID": "your_razorpay_key_id",
                "RAZORPAY_KEY_SECRET": "your_razorpay_key_secret"
            }
        }
    }
}
```
Please replace the `your_razorpay_key_id` and `your_razorpay_key_secret` with your keys.

- Learn about how to configure MCP servers in Claude desktop: [Link](https://modelcontextprotocol.io/quickstart/user)
- How to install Claude Desktop: [Link](https://claude.ai/download)

## Usage with VS Code

Add the following to your VS Code settings (JSON):

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "razorpay_key_id",
        "description": "Razorpay Key ID",
        "password": false
      },
      {
        "type": "promptString",
        "id": "razorpay_key_secret",
        "description": "Razorpay Key Secret",
        "password": true
      }
    ],
    "servers": {
      "razorpay": {
        "command": "docker",
        "args": [
          "run",
          "-i",
          "--rm",
          "-e",
          "RAZORPAY_KEY_ID",
          "-e",
          "RAZORPAY_KEY_SECRET",
          "razorpay-mcp-server:latest"
        ],
        "env": {
          "RAZORPAY_KEY_ID": "${input:razorpay_key_id}",
          "RAZORPAY_KEY_SECRET": "${input:razorpay_key_secret}"
        }
      }
    }
  }
}
```

Learn more about MCP servers in VS Code's [agent mode documentation](https://code.visualstudio.com/docs/copilot/chat/mcp-servers).

## Configuration

The server requires the following configuration:

- `RAZORPAY_KEY_ID`: Your Razorpay API key ID
- `RAZORPAY_KEY_SECRET`: Your Razorpay API key secret
- `LOG_FILE` (optional): Path to log file for server logs
- `TOOLSETS` (optional): Comma-separated list of toolsets to enable (default: "all")
- `READ_ONLY` (optional): Run server in read-only mode (default: false)

### Command Line Flags

The server supports the following command line flags:

- `--key` or `-k`: Your Razorpay API key ID
- `--secret` or `-s`: Your Razorpay API key secret
- `--log-file` or `-l`: Path to log file
- `--toolsets` or `-t`: Comma-separated list of toolsets to enable
- `--read-only`: Run server in read-only mode

## Debugging the Server

You can use the standard Go debugging tools to troubleshoot issues with the server. Log files can be specified using the `--log-file` flag (defaults to ./logs)

## License

This project is licensed under the terms of the MIT open source license. Please refer to [LICENSE](./LICENSE) for the full terms.
