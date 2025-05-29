# Razorpay MCP Server (Official)

The Razorpay MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with Razorpay APIs, enabling advanced payment processing capabilities for developers and AI tools.

## Available Tools

Currently, the Razorpay MCP Server provides the following tools:

| Tool                                 | Description                                            | API
|:-------------------------------------|:-------------------------------------------------------|:-----------------------------------
| `capture_payment`                    | Change the payment status from authorized to captured. | [Payment](https://razorpay.com/docs/api/payments/capture)
| `fetch_payment`                      | Fetch payment details with ID                          | [Payment](https://razorpay.com/docs/api/payments/fetch-with-id)
| `fetch_payment_card_details`         | Fetch card details used for a payment                  | [Payment](https://razorpay.com/docs/api/payments/fetch-payment-expanded-card)
| `fetch_all_payments`                 | Fetch all payments with filtering and pagination       | [Payment](https://razorpay.com/docs/api/payments/fetch-all-payments)
| `update_payment`                     | Update the notes field of a payment                    | [Payment](https://razorpay.com/docs/api/payments/update)|
| `create_payment_link`                | Creates a new payment link (standard)                  | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/create-standard)
| `create_payment_link_upi`            | Creates a new UPI payment link                         | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/create-upi)
| `fetch_all_payment_links`            | Fetch all the payment links                            | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/fetch-all-standard)
| `fetch_payment_link`                 | Fetch details of a payment link                        | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/fetch-id-standard/)
| `send_payment_link`                  | Send a payment link via SMS or email.                  | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/resend)
| `update_payment_link`                | Updates a new standard payment link                    | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/update-standard)
| `create_order`                       | Creates an order                                       | [Order](https://razorpay.com/docs/api/orders/create/)
| `fetch_order`                        | Fetch order with ID                                    | [Order](https://razorpay.com/docs/api/orders/fetch-with-id)
| `fetch_all_orders`                   | Fetch all orders                                       | [Order](https://razorpay.com/docs/api/orders/fetch-all)
| `update_order`                       | Update an order                                        | [Order](https://razorpay.com/docs/api/orders/update) 
| `fetch_order_payments`               | Fetch all payments for an order                        | [Order](https://razorpay.com/docs/api/orders/fetch-payments/)
| `create_refund`                      | Creates a refund                                       | [Refund](https://razorpay.com/docs/api/refunds/create-instant/)
| `fetch_refund`                       | Fetch refund details with ID                           | [Refund](https://razorpay.com/docs/api/refunds/fetch-with-id/)
| `fetch_all_refunds`                  | Fetch all refunds                                      | [Refund](https://razorpay.com/docs/api/refunds/fetch-all)
| `update_refund`                      | Update refund notes with ID                            | [Refund](https://razorpay.com/docs/api/refunds/update/)
| `fetch_multiple_refunds_for_payment` | Fetch multiple refunds for a payment                   | [Refund](https://razorpay.com/docs/api/refunds/fetch-multiple-refund-payment/)
| `fetch_specific_refund_for_payment`  | Fetch a specific refund for a payment                  | [Refund](https://razorpay.com/docs/api/refunds/fetch-specific-refund-payment/)
| `create_qr_code`                     | Creates a QR Code                                      | [QR Code](https://razorpay.com/docs/api/qr-codes/create/)
| `fetch_qr_code`                      | Fetch QR Code with ID                                  | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-with-id/)
| `fetch_all_qr_codes`                 | Fetch all QR Codes                                     | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-all/)
| `fetch_qr_codes_by_customer_id`      | Fetch QR Codes with Customer ID                        | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-customer-id/)
| `fetch_qr_codes_by_payment_id`       | Fetch QR Codes with Payment ID                         | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-payment-id/)
| `fetch_payments_for_qr_code`         | Fetch Payments for a QR Code                           | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-payments/)
| `close_qr_code`                      | Closes a QR Code                                       | [QR Code](https://razorpay.com/docs/api/qr-codes/close/)
| `fetch_all_settlements`              | Fetch all settlements                                  | [Settlement](https://razorpay.com/docs/api/settlements/fetch-all)
| `fetch_settlement_with_id`           | Fetch settlement details                               | [Settlement](https://razorpay.com/docs/api/settlements/fetch-with-id)
| `fetch_settlement_recon_details`     | Fetch settlement reconciliation report                 | [Settlement](https://razorpay.com/docs/api/settlements/fetch-recon)
| `create_instant_settlement`          | Create an instant settlement                           | [Settlement](https://razorpay.com/docs/api/settlements/instant/create)
| `fetch_all_instant_settlements`      | Fetch all instant settlements                          | [Settlement](https://razorpay.com/docs/api/settlements/instant/fetch-all)
| `fetch_instant_settlement_with_id`   | Fetch instant settlement with ID                       | [Settlement](https://razorpay.com/docs/api/settlements/instant/fetch-with-id)
| `fetch_all_payouts`                  | Fetch all payout details with A/c number               | [Payout](https://razorpay.com/docs/api/x/payouts/fetch-all/)
| `fetch_payout_by_id`                 | Fetch the payout details with payout ID                | [Payout](https://razorpay.com/docs/api/x/payouts/fetch-with-id)


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

## Usage with SSE Server

The Razorpay MCP Server also supports the Server-Sent Events (SSE) transport protocol, allowing you to run it as a standalone service that clients can connect to.

### Running the SSE Server with Docker

To run the server in SSE mode using the same Docker image:

```bash
# Run the SSE server on port 8090 (default)
docker run -p 8090:8090 \
  -e MODE=sse \
  -e PORT=8090 \
  razorpay-mcp-server:latest
```

You can customize the port by setting the `PORT` environment variable.

### Testing with MCP Inspector

You can test your SSE server using the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) tool:

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Open MCP Inspector tool
npx @modelcontextprotocol/inspector
```

This will open a browser interface where you can inspect and test the available tools on your SSE server.

## Configuration

The server requires the following configuration:

- `RAZORPAY_KEY_ID`: Your Razorpay API key ID (required for stdio mode)
- `RAZORPAY_KEY_SECRET`: Your Razorpay API key secret (required for stdio mode)
- `MODE`: Server mode ("stdio" or "sse", default: "stdio")
- `PORT`: Port for SSE server (default: "8090", used in SSE mode)
- `LOG_FILE` (optional): Path to log file for stdio mode logs (SSE mode always logs to stdout)
- `TOOLSETS` (optional): Comma-separated list of toolsets to enable (default: "all")
- `READ_ONLY` (optional): Run server in read-only mode (default: false)

### Command Line Flags

The server supports different flags based on the mode:

For stdio mode:
- `--key` or `-k`: Your Razorpay API key ID
- `--secret` or `-s`: Your Razorpay API key secret
- `--log-file` or `-l`: Path to log file
- `--toolsets` or `-t`: Comma-separated list of toolsets to enable
- `--read-only`: Run server in read-only mode

For SSE mode:
- `--port`: Port to run the SSE server on (default: 8090)
- `--toolsets` or `-t`: Comma-separated list of toolsets to enable
- `--read-only`: Run server in read-only mode
Note: SSE mode logs are always written to stdout for better container integration

## Debugging the Server

You can use the standard Go debugging tools to troubleshoot issues with the server. Log files can be specified using the `--log-file` flag (defaults to ./logs)

## License

This project is licensed under the terms of the MIT open source license. Please refer to [LICENSE](./LICENSE) for the full terms.
