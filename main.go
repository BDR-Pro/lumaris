package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bdr-pro/lumaris/auth"
	"github.com/bdr-pro/lumaris/buyer"
	"github.com/bdr-pro/lumaris/seller"
)

func main() {
	// Define our own flagset so we can parse our top-level commands
	mainFlags := flag.NewFlagSet("main", flag.ExitOnError)

	// Help text
	mainFlags.Usage = func() {
		fmt.Println("Lumaris Compute Marketplace CLI")
		fmt.Println("\nUsage:")
		fmt.Println("  main [mode] [options]")
		fmt.Println("\nModes:")
		fmt.Println("  buyer    - Run as a buyer to submit compute jobs")
		fmt.Println("  seller   - Run as a seller to execute compute jobs")
		fmt.Println("  test-buy - Test the buyer functionality")
		fmt.Println("  test-sell - Test the seller functionality")
		fmt.Println("  auth     - Authenticate with Nakama server and get a valid token")
		fmt.Println("\nOptions:")
		fmt.Println("  -h, --help    - Show this help message")
		fmt.Println("\nFor mode-specific options, run: main [mode] --help")
	}

	// Check if we have any arguments
	if len(os.Args) < 2 {
		mainFlags.Usage()
		os.Exit(1)
	}

	// Parse the mode
	switch os.Args[1] {
	case "buyer":
		// Start the buyer client
		buyer.ClientMain()
	case "seller":
		// Start the seller runner
		seller.RunnerMain()
	case "test-buy":
		// Run buyer test
		buyer.Test()
	case "test-sell":
		// Run seller test - not implemented yet
		fmt.Println("Seller test not implemented yet.")
		os.Exit(1)
	case "auth":
		// Run authentication
		handleAuth()
	case "-h", "--help":
		mainFlags.Usage()
	default:
		fmt.Printf("Unknown mode: %s\n", os.Args[1])
		mainFlags.Usage()
		os.Exit(1)
	}
}

// Ensure all the function signatures are included so they can be called from main
// These functions are defined in their respective files:
// - buyer/buyerClient.go
// - buyer/buyerTest.go
// - seller/sellerRunner.go

// No need for these placeholder variables anymore since we're
// using proper imports now
func handleAuth() {
	authFlags := flag.NewFlagSet("auth", flag.ExitOnError)
	server := authFlags.String("server", "127.0.0.1:7350", "Nakama server address")
	method := authFlags.String("method", "server", "Authentication method: email, device, or server")
	email := authFlags.String("email", "", "Email for email authentication")
	password := authFlags.String("password", "", "Password for email authentication")
	deviceID := authFlags.String("device", "", "Device ID for device authentication")
	createAccount := authFlags.Bool("create", false, "Create account if it doesn't exist")
	serverKeytemp := authFlags.String("key", "", "Server key for server authentication")

	if err := authFlags.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Failed to parse auth flags: %v", err)
	}

	serverKeyRaw := *serverKeytemp + ":"
	serverKey := base64.StdEncoding.EncodeToString([]byte(serverKeyRaw))

	var authResp *auth.NakamaAuthResponse
	var err error

	switch *method {
	case "email":
		if *email == "" || *password == "" {
			log.Fatal("Email authentication requires both -email and -password flags")
		}
		authResp, err = auth.AuthenticateWithEmail(*server, *email, *password, *createAccount)
	case "device":
		if *deviceID == "" {
			log.Fatal("Device authentication requires -device flag")
		}
		authResp, err = auth.AuthenticateWithDeviceID(*server, *deviceID, *createAccount)
	case "server":
		if *serverKeytemp == "" {
			log.Fatal("Server authentication requires -key flag")
		}
		authResp, err = auth.AuthenticateWithServerKey(*server, serverKey)
	default:
		log.Fatalf("Unknown authentication method: %s", *method)
	}

	if err != nil {
		log.Printf("Authentication failed: %v", err)
		log.Println("\nPossible solutions:")
		log.Println("1. Check if your server address is correct")
		log.Println("2. Verify you're using the correct server key")
		log.Println("3. Make sure the Nakama server is running")
		log.Println("4. Try using a different authentication method")
		log.Println("\nFor debug purposes, try creating a user with:")
		log.Printf("  ./lumaris auth -server %s -method email -email test@example.com -password password -create\n", *server)
		os.Exit(1)
	}

	fmt.Println("✅ Authentication successful!")
	fmt.Println("🔐 Session token:", authResp.Token)
	fmt.Println("👤 User ID:", authResp.UserID)
	fmt.Println()
	fmt.Println("To use this token, run commands like:")
	fmt.Printf("  ./lumaris buyer -server %s -token %s\n", *server, authResp.Token)
}
