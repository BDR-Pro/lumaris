# Lumaris Compute Marketplace

A distributed marketplace for computing resources built on Nakama game server.

## Project Structure

```bash
lumaris/
├── main.go              # Main application entry point
├── modules/
│   ├── jobReq.go        # Job request/result data structures
│   └── nakamaModule.go  # Nakama server-side module code
├── buyer/
│   ├── client.go        # Buyer client implementation
│   └── test.go          # Buyer test implementation
└── seller/
    └── runner.go        # Seller runner implementation
```

## Usage

### Building the project

```bash
go build -o lumaris
```

### Running as a buyer

```bash
./lumaris buyer -server 127.0.0.1:7350 -token your_token_here
```

### Running as a seller

```bash
./lumaris seller -server 127.0.0.1:7350 -token your_token_here
```

### Running tests

Test the buyer functionality:

```bash
./lumaris test-buy -server 127.0.0.1:7350 -token your_token_here -image python:3.10 -command "python -c 'print(\"Hello\")"
```

## Command-line Options

### Global Options

- `-server` - Nakama server address (default: 127.0.0.1:7350)
- `-token` - Nakama session token (required)

### Buyer Test Options

- `-image` - Docker image to use (default: python:3.10)
- `-command` - Command to run inside the container

## Development

### Nakama Module Setup

The Nakama module (`modules/nakamaModule.go`) needs to be compiled and loaded into your Nakama server:

1. Build the module:

   ```bash
   go build -buildmode=plugin -o ./modules.so ./modules/nakamaModule.go
   ```

2. Add the module to your Nakama configuration:

   ```yaml
   runtime:
     path: "/path/to/modules.so"
   ```

3. Restart Nakama server to load the module.
