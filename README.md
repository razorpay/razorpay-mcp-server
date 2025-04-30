# Razorpay MCP Server (Official)

The Razorpay MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with Razorpay APIs, enabling advanced payment processing capabilities for developers and AI tools.

## Available Tools

Currently, the Razorpay MCP Server provides the following tools:

| Tool                  | Description                                     | API
|:----------------------|:------------------------------------------------|:-----------------------------------
| `fetch_payment`       | Fetch payment details with ID                   | [Payment](https://razorpay.com/docs/api/payments/fetch-with-id)
| `create_payment_link` | Creates a new payment link (standard)           | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/create-standard)
| `fetch_payment_link`  | Fetch details of a payment link (standard)      | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/fetch-id-standard/)
| `create_order`        | Creates an order                                | [Order](https://razorpay.com/docs/api/orders/create/)
| `fetch_order`         | Fetch order with ID                             | [Order](https://razorpay.com/docs/api/orders/fetch-with-id)
| `fetch_all_orders`    | Fetch all orders                                | [Order](https://razorpay.com/docs/api/orders/fetch-all)
| `create_refund`       | Creates a refund                                | [Refund](https://razorpay.com/docs/api/refunds/create-instant/)
| `fetch_refund`        | Fetch refund details with ID                    | [Refund](https://razorpay.com/docs/api/refunds/fetch-with-id/)
| `update_refund`       | Update refund notes with ID                     | [Refund](https://razorpay.com/docs/api/refunds/update/)

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
