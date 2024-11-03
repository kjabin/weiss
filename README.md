# Weiss

A simple WebSocket test client written in Go, designed to make WebSocket debugging easier with color-formatted output.

## Installation

```bash
go install github.com/kjabin/weiss
```

## Usage

```bash
weiss [URL]
```

## Message Format

When receiving messages from the server, Weiss displays:
{ Kind: **Message Type**, Message: **Message Content** }

Messages are color-formatted for better visibility and debugging experience.
