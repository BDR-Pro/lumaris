# LUMARIS: Decentralized Compute Marketplace

LUMARIS is a decentralized compute marketplace built on top of Nakama and Docker. It enables buyers to send compute jobs to sellers who execute these jobs in Docker containers and return the results.

## System Architecture

The system consists of three main components:

1. **Buyers**: Applications that need compute resources and send job requests
2. **Sellers**: Nodes that provide compute resources by running Docker containers
3. **Nakama Server**: Coordinates communication between buyers and sellers

## Prerequisites

- Go 1.16 or newer
- Nakama server running (local or remote)
- Docker installed on seller machines

## Project Structure

```bash
LUMARIS/
├── buyer/
│   ├── buyerClient.go    # Sends job requests and receives results
│   └── buyerSample.go    # Simple one-time job sender
├── seller/
│   └── sellerRunner.go   # Listens for jobs and runs Docker containers
├── modules/
│   ├── jobReq.go         # Job request/result data structures
│   └── nakamaModule.go   # Nakama RPC handlers
├── go.mod
└── go.sum
```

## How It Works

1. **Seller Registration**: Sellers register with the Nakama server by joining a topic and declaring their capabilities.
2. **Job Submission**: Buyers send job requests (Docker image + command) to Nakama.
3. **Job Distribution**: Nakama forwards the job to available sellers.
4. **Job Execution**: Sellers run the job in a Docker container with security constraints.
5. **Result Delivery**: Sellers send the job results back to the buyer via Nakama.

## Security Considerations

- Docker containers run with network disabled (`--network=none`)
- Memory and CPU limits are enforced (`--memory=512m --cpus=1`)
- Each job runs in an isolated container

## Setup Instructions

### 1. Set up Nakama Server

First, you need to set up a Nakama server. You can run it locally or use a hosted service.

For local development, you can use Docker:

```bash
docker run --name nakama -p 7350:7350 -p 7349:7349 -d heroiclabs/nakama:latest
```

### 2. Deploy the Nakama Module

Copy the files from the `modules` directory to your Nakama server's modules directory.

### 3. Start Sellers

Start one or more seller instances:

```bash
go run seller/sellerRunner.go --server 127.0.0.1:7350 --key defaultkey
```

### 4. Submit Jobs

You can submit jobs in two ways:

- One-time job submission:

  ```bash
  go run buyer/buyerSample.go --server 127.0.0.1:7350 --key defaultkey --image python:3.10 --command "python -c 'print(\"Hello world!\")'"
  ```

- Interactive client (receives results):

  ```bash
  go run buyer/buyerClient.go --server 127.0.0.1:7350 --key defaultkey
  ```

## Extending the System

### Adding Payment Processing

For a complete marketplace, you would want to add payment processing. This could be done by:

1. Adding payment details to the job request
2. Implementing transaction verification in the Nakama module
3. Integrating with a payment processor or cryptocurrency

### Improving Job Matching

The current system broadcasts job requests to all sellers. You could improve it by:

1. Implementing a more sophisticated matching algorithm
2. Adding seller ratings and job success statistics
3. Creating a bidding system where sellers compete for jobs

### Enhanced Security

Consider implementing:

1. Image whitelisting for allowed Docker images
2. Reputation system for sellers
3. Runtime monitoring of containers

## Troubleshooting

- **Docker Permission Issues**: Make sure the user running the seller has permissions to run Docker commands. You may need to add the user to the docker group: `sudo usermod -aG docker $USER`
- **Nakama Connection Issues**: Verify that the Nakama server is running and accessible from your network.
- **Job Not Being Executed**: Check that at least one seller is registered and active.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
