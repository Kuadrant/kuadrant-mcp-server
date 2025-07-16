#!/bin/bash
# Run the MCP server with logging to both stderr and a file

LOG_FILE="/tmp/kuadrant-mcp.log"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Starting Kuadrant MCP server with logging to $LOG_FILE" >&2
echo "You can monitor the log with: tail -f $LOG_FILE" >&2

# Run the server, tee stderr to log file while preserving stdio for MCP
exec "$SCRIPT_DIR/kuadrant-mcp-server" 2> >(tee -a "$LOG_FILE" >&2)