# Razorpay MCP Server

The Razorpay MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with Razorpay APIs, enabling advanced payment processing capabilities for developers and AI tools.

## Use Cases 
Bring Razorpay to your agentic applications using Razorpay MCP Server.

- Agentic Applications: Building AI powered tools that interact with Razorpay's payment ecosystem using this Razorpay MCP server.
- Analytics Usecases: Fetching payment data from Razorpay for analysis or customer support.
- Customer and Operational Usecases: You can bring Razorpay integration into your agentic customer and operational dashboards using Razorpay MCP server.

## Setup

To run the Razorpay MCP server, use one of the following methods:

### Using Docker (Recommended)

```bash
# Run the server
docker run -i --rm \
  -e RAZORPAY_KEY_ID=your_key_id \
  -e RAZORPAY_KEY_SECRET=your_key_secret \
  <TODO>/razorpay/razorpay-mcp-server
```

Replace `your_key_id` and `your_key_secret` with your actual Razorpay API credentials.

### Build from source

```bash
# Clone the repository
git clone https://github.com/razorpay/razorpay-mcp-server.git
cd razorpay-mcp-server

# Build the binary
go build -o razorpay-mcp-server ./cmd/razorpay-mcp-server

# Run the server
RAZORPAY_KEY_ID=your_key_id RAZORPAY_KEY_SECRET=your_key_secret ./razorpay-mcp-server stdio
```

## Usage with Razorpay Checkout
Coming soon.

## Usage with Claude Desktop

Add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
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
        "<TODO>/razorpay/razorpay-mcp-server"
      ],
      "env": {
        "RAZORPAY_KEY_ID": "your_key_id",
        "RAZORPAY_KEY_SECRET": "your_key_secret"
      }
    }
  }
}
```

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
          "<TODO>/razorpay/razorpay-mcp-server"
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

## Available Tools

Currently, the Razorpay MCP Server provides the following tools:

| Tool                  | Description                           |
|-----------------------|---------------------------------------|
| `payment.fetch`       | Fetch payment details                 |
| `payment_link.create` | Creates a new payment link            |
| `payment_link.fetch`  | Fetch details of a payment link       |
| `order.create`        | Creates an order                      |
| `order.fetch`         | Fetch order details                   |

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
