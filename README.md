# Razorpay MCP Server (Official)

The Razorpay MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with Razorpay APIs, enabling advanced payment processing capabilities for developers and AI tools.

## Quick Start

Choose your preferred setup method:
- **[Remote MCP Server](#remote-mcp-server-recommended)** - Hosted by Razorpay, no setup required
- **[Local MCP Server](#local-mcp-server)** - Run on your own infrastructure

## Available Tools

Currently, the Razorpay MCP Server provides the following tools:

| Tool                                 | Description                                            | API | Remote Server Support |
|:-------------------------------------|:-------------------------------------------------------|:------------------------------------|:---------------------|
| `capture_payment`                    | Change the payment status from authorized to captured. | [Payment](https://razorpay.com/docs/api/payments/capture) | ✅ |
| `fetch_payment`                      | Fetch payment details with ID                          | [Payment](https://razorpay.com/docs/api/payments/fetch-with-id) | ✅ |
| `fetch_payment_card_details`         | Fetch card details used for a payment                  | [Payment](https://razorpay.com/docs/api/payments/fetch-payment-expanded-card) | ✅ |
| `fetch_all_payments`                 | Fetch all payments with filtering and pagination       | [Payment](https://razorpay.com/docs/api/payments/fetch-all-payments) | ✅ |
| `update_payment`                     | Update the notes field of a payment                    | [Payment](https://razorpay.com/docs/api/payments/update) | ✅ |
| `create_payment_link`                | Creates a new payment link (standard)                  | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/create-standard) | ✅ |
| `create_payment_link_upi`            | Creates a new UPI payment link                         | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/create-upi) | ✅ |
| `fetch_all_payment_links`            | Fetch all the payment links                            | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/fetch-all-standard) | ✅ |
| `fetch_payment_link`                 | Fetch details of a payment link                        | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/fetch-id-standard/) | ✅ |
| `send_payment_link`                  | Send a payment link via SMS or email.                  | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/resend) | ✅ |
| `update_payment_link`                | Updates a new standard payment link                    | [Payment Link](https://razorpay.com/docs/api/payments/payment-links/update-standard) | ✅ |
| `create_order`                       | Creates an order                                       | [Order](https://razorpay.com/docs/api/orders/create/) | ✅ |
| `fetch_order`                        | Fetch order with ID                                    | [Order](https://razorpay.com/docs/api/orders/fetch-with-id) | ✅ |
| `fetch_all_orders`                   | Fetch all orders                                       | [Order](https://razorpay.com/docs/api/orders/fetch-all) | ✅ |
| `update_order`                       | Update an order                                        | [Order](https://razorpay.com/docs/api/orders/update) | ✅ |
| `fetch_order_payments`               | Fetch all payments for an order                        | [Order](https://razorpay.com/docs/api/orders/fetch-payments/) | ✅ |
| `create_refund`                      | Creates a refund                                       | [Refund](https://razorpay.com/docs/api/refunds/create-instant/) | ❌ |
| `fetch_refund`                       | Fetch refund details with ID                           | [Refund](https://razorpay.com/docs/api/refunds/fetch-with-id/) | ✅ |
| `fetch_all_refunds`                  | Fetch all refunds                                      | [Refund](https://razorpay.com/docs/api/refunds/fetch-all) | ✅ |
| `update_refund`                      | Update refund notes with ID                            | [Refund](https://razorpay.com/docs/api/refunds/update/) | ✅ |
| `fetch_multiple_refunds_for_payment` | Fetch multiple refunds for a payment                   | [Refund](https://razorpay.com/docs/api/refunds/fetch-multiple-refund-payment/) | ✅ |
| `fetch_specific_refund_for_payment`  | Fetch a specific refund for a payment                  | [Refund](https://razorpay.com/docs/api/refunds/fetch-specific-refund-payment/) | ✅ |
| `create_qr_code`                     | Creates a QR Code                                      | [QR Code](https://razorpay.com/docs/api/qr-codes/create/) | ✅ |
| `fetch_qr_code`                      | Fetch QR Code with ID                                  | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-with-id/) | ✅ |
| `fetch_all_qr_codes`                 | Fetch all QR Codes                                     | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-all/) | ✅ |
| `fetch_qr_codes_by_customer_id`      | Fetch QR Codes with Customer ID                        | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-customer-id/) | ✅ |
| `fetch_qr_codes_by_payment_id`       | Fetch QR Codes with Payment ID                         | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-payment-id/) | ✅ |
| `fetch_payments_for_qr_code`         | Fetch Payments for a QR Code                           | [QR Code](https://razorpay.com/docs/api/qr-codes/fetch-payments/) | ✅ |
| `close_qr_code`                      | Closes a QR Code                                       | [QR Code](https://razorpay.com/docs/api/qr-codes/close/) | ❌ |
| `fetch_all_settlements`              | Fetch all settlements                                  | [Settlement](https://razorpay.com/docs/api/settlements/fetch-all) | ✅ |
| `fetch_settlement_with_id`           | Fetch settlement details                               | [Settlement](https://razorpay.com/docs/api/settlements/fetch-with-id) | ✅ |
| `fetch_settlement_recon_details`     | Fetch settlement reconciliation report                 | [Settlement](https://razorpay.com/docs/api/settlements/fetch-recon) | ✅ |
| `create_instant_settlement`          | Create an instant settlement                           | [Settlement](https://razorpay.com/docs/api/settlements/instant/create) | ❌ |
| `fetch_all_instant_settlements`      | Fetch all instant settlements                          | [Settlement](https://razorpay.com/docs/api/settlements/instant/fetch-all) | ✅ |
| `fetch_instant_settlement_with_id`   | Fetch instant settlement with ID                       | [Settlement](https://razorpay.com/docs/api/settlements/instant/fetch-with-id) | ✅ |
| `fetch_all_payouts`                  | Fetch all payout details with A/c number               | [Payout](https://razorpay.com/docs/api/x/payouts/fetch-all/) | ✅ |
| `fetch_payout_by_id`                 | Fetch the payout details with payout ID                | [Payout](https://razorpay.com/docs/api/x/payouts/fetch-with-id) | ✅ |


## Use Cases
- Workflow Automation: Automate your day to day workflow using Razorpay MCP Server.
- Agentic Applications: Building AI powered tools that interact with Razorpay's payment ecosystem using this Razorpay MCP server.

## Remote MCP Server (Recommended)

The Remote MCP Server is hosted by Razorpay and provides instant access to Razorpay APIs without any local setup. This is the recommended approach for most users.

### Benefits of Remote MCP Server

- **Zero Setup**: No need to install Docker, Go, or manage local infrastructure
- **Always Updated**: Automatically stays updated with the latest features and security patches
- **High Availability**: Backed by Razorpay's robust infrastructure with 99.9% uptime
- **Reduced Latency**: Optimized routing and caching for faster API responses
- **Enhanced Security**: Secure token-based authentication with automatic token rotation
- **No Maintenance**: No need to worry about updates, patches, or server maintenance

### Prerequisites

`npx` is needed to use mcp server.
You need to have Node.js installed on your system, which includes both `npm` (Node Package Manager) and `npx` (Node Package Execute) by default:

#### macOS
```bash
# Install Node.js (which includes npm and npx) using Homebrew
brew install node

# Alternatively, download from https://nodejs.org/
```

#### Windows
```bash
# Install Node.js (which includes npm and npx) using Chocolatey
choco install nodejs

# Alternatively, download from https://nodejs.org/
```

#### Verify Installation
```bash
npx --version
```

### Usage with Cursor

Inside your cursor settings in MCP, add this config.

```json
{
  "mcpServers": {
    "rzp-sse-mcp-server": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.razorpay.com/sse",
        "--header",
        "Authorization:${AUTH_HEADER}"
      ],
      "env": {
        "AUTH_HEADER": "Bearer <Base64(key:secret)>"
      }
    }
  }
}
```

Replace `key` & `secret` with your Razorpay API KEY & API SECRET

### Usage with Claude Desktop

Add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "rzp-sse-mcp-server": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.razorpay.com/sse",
        "--header",
        "Authorization: Bearer <Merchant Token>"
      ]
    }
  }
}
```

Replace `<Merchant Token>` with your Razorpay merchant token. Check Authentication section for steps to generate token.

- Learn about how to configure MCP servers in Claude desktop: [Link](https://modelcontextprotocol.io/quickstart/user)
- How to install Claude Desktop: [Link](https://claude.ai/download)

### Usage with VS Code

Add the following to your VS Code settings (JSON):

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "merchant_token",
        "description": "Razorpay Merchant Token",
        "password": true
      }
    ],
    "servers": {
      "razorpay-remote": {
        "command": "npx",
        "args": [
          "mcp-remote",
          "https://mcp.razorpay.com/sse",
          "--header",
          "Authorization: Bearer ${input:merchant_token}"
        ]
      }
    }
  }
}
```

Learn more about MCP servers in VS Code's [agent mode documentation](https://code.visualstudio.com/docs/copilot/chat/mcp-servers).


## Authentication

The Remote MCP Server uses merchant token-based authentication. To generate your merchant token:

1. Go to the [Razorpay Dashboard](https://dashboard.razorpay.com/) and navigate to Settings > API Keys
2. Locate your API Key and API Secret:
   - API Key is visible on the dashboard
   - API Secret is generated only once when you first create it. **Important:** Do not generate a new secret if you already have one

3. Generate your merchant token by running this command in your terminal:
   ```bash
   echo <RAZORPAY_API_KEY>:<RAZORPAY_API_SECRET> | base64
   ```
   Replace `<RAZORPAY_API_KEY>` and `<RAZORPAY_API_SECRET>` with your actual credentials

4. Copy the base64-encoded output - this is your merchant token for the Remote MCP Server

> **Note:** For local MCP Server deployment, you can use the API Key and Secret directly without generating a merchant token.
     

## Local MCP Server

For users who prefer to run the MCP server on their own infrastructure or need access to all tools (including those restricted in the remote server), you can deploy the server locally.

### Prerequisites

- Docker
- Golang (Go)
- Git

To run the Razorpay MCP server, use one of the following methods:

### Using Public Docker Image (Recommended)

You can use the public Razorpay image directly. No need to build anything yourself - just copy-paste the configurations below and make sure Docker is already installed.

> **Note:** To use a specific version instead of the latest, replace `razorpay/mcp` with `razorpay/mcp:v1.0.0` (or your desired version tag) in the configurations below. Available tags can be found on [Docker Hub](https://hub.docker.com/r/razorpay/mcp/tags).


#### Usage with Claude Desktop

This will use the public razorpay image

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
                "razorpay/mcp"
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

#### Usage with VS Code

Add the following to your VS Code settings (JSON):

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
                "razorpay/mcp"
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

#### Usage with VS Code

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
          "razorpay/mcp"
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

### Build from Docker (Alternative)

You need to clone the Github repo and build the image for Razorpay MCP Server using `docker`. Do make sure `docker` is installed and running in your system.

```bash
# Run the server
git clone https://github.com/razorpay/razorpay-mcp-server.git
cd razorpay-mcp-server
docker build -t razorpay-mcp-server:latest .
```

Once the razorpay-mcp-server:latest docker image is built, you can replace the public image(`razorpay/mcp`) with it in the above configurations.

### Build from source

You can directly build from the source instead of using docker by following these steps:

```bash
# Clone the repository
git clone https://github.com/razorpay/razorpay-mcp-server.git
cd razorpay-mcp-server

# Build the binary
go build -o razorpay-mcp-server ./cmd/razorpay-mcp-server
```
Once the build is ready, you need to specify the path to the binary executable in the `command` option. Here's an example for VS Code settings:

```json
{
  "razorpay": {
    "command": "/path/to/razorpay-mcp-server",
    "args": ["stdio","--log-file=/path/to/rzp-mcp.log"],
    "env": {
      "RAZORPAY_KEY_ID": "<YOUR_ID>",
      "RAZORPAY_KEY_SECRET" : "<YOUR_SECRET>"
    }
  }
}
```

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