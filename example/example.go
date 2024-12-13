package main

import (
	"fmt"
	"log"
	"time"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/generated/clients/fundingsdk"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/generated/clients/holidayssdk"
)

// The ExampleSecuritySource has been removed since our client now filters out external endpoints
// and provides a NewInternalClient function that doesn't require authentication

func main() {
	fmt.Println("=== Example of using generated internal clients ===")
	fmt.Println("Note: Using example URLs - in a real environment, use actual service URLs")
	fmt.Println()

	// Example 1: Initialize holidays client (internal endpoints only)
	fmt.Println("=== Holidays Client Example ===")
	// Using the NewInternalClient function which doesn't require a security source
	fmt.Println("Initializing holidays client with NewInternalClient (no security source needed)")
	holidaysClient, err := holidayssdk.NewInternalClient("https://holidays-server.example.com")
	if err != nil {
		log.Fatalf("Failed to create holidays client: %v", err)
	}
	fmt.Printf("Successfully created holidaysClient pointing to: %s\n",
		"https://holidays-server.example.com")

	// Since we're using a dummy URL, we'll get errors when trying to call real endpoints
	// In a real application, you would use actual service URLs

	// Example 1: Check if a date is a holiday
	fmt.Println("\nExample 1: Check if a date is a holiday")
	today := time.Now().Format("2006-01-02")
	fmt.Printf("Checking if today (%s) is a holiday...\n", today)

	// In real usage, you would call this and handle errors:
	/*
		isHolidayResp, err := holidaysClient.HolidaysInternalControllerIsHoliday(context.Background(), holidayssdk.HolidaysInternalControllerIsHolidayParams{
			Date: today,
		})
	*/

	// For the example, we'll show mock data:
	fmt.Printf("Is today (%s) a holiday? false (mock response)\n", today)

	// Example 2: Get upcoming holidays (internal endpoint)
	fmt.Println("\nExample 2: Get upcoming holidays")
	fmt.Printf("Getting holidays from %s for the next 30 days...\n", today)

	// In real usage:
	/*
		upcomingHolidays, err := holidaysClient.HolidaysInternalControllerGetUpcomingHolidays(context.Background(), holidayssdk.HolidaysInternalControllerGetUpcomingHolidaysParams{
			FromDate:   today,
			DaysLookup: 30,
		})
	*/

	// For the example, show mock data:
	fmt.Println("Upcoming holidays (mock response):")
	fmt.Println("- 2025-03-01: Weekend")
	fmt.Println("- 2025-03-02: Weekend")
	fmt.Println("- 2025-04-18: Good Friday")

	// Example 3: Initialize funding client for internal endpoints only
	fmt.Println("\n=== Funding Client Example ===")
	// Using the NewInternalClient function which doesn't require a security source
	fmt.Println("Initializing funding client with NewInternalClient (no security source needed)")
	fundingClient, err := fundingsdk.NewInternalClient("https://funding-server.example.com")
	if err != nil {
		log.Fatalf("Failed to create funding client: %v", err)
	}
	fmt.Printf("Successfully created fundingClient pointing to: %s\n",
		"https://funding-server.example.com")

	// Example 3: Get deposit records for an account (internal endpoint)
	fmt.Println("\nExample 3: Get deposit records for an account")
	accountUUID := "example-account-uuid"
	fmt.Printf("Getting deposit records for account %s...\n", accountUUID)

	// In real usage:
	/*
		depositRecords, err := fundingClient.DepositRecordsInternalControllerGetDepositRecordsByAccount(context.Background(), fundingsdk.DepositRecordsInternalControllerGetDepositRecordsByAccountParams{
			AccountUuid: accountUUID,
		})
	*/

	// For the example, show mock data:
	fmt.Printf("Deposit records for account %s (mock response):\n", accountUUID)
	fmt.Println("- Amount: 1000.00, Status: COMPLETED")
	fmt.Println("- Amount: 500.00, Status: PENDING")

	fmt.Println("\n=== Client Configuration Examples ===")
	fmt.Println("// Holidays client (no security source needed):")
	fmt.Println("holidaysClient, err := holidayssdk.NewInternalClient(\"https://holidays-server.example.com\")")

	fmt.Println("\n// Funding client for internal endpoints (no security source needed):")
	fmt.Println("fundingClient, err := fundingsdk.NewInternalClient(\"https://funding-server.example.com\")")

	fmt.Println("\nNote: All SDK clients now provide NewInternalClient for accessing internal endpoints without security.")

	// Use the clients to avoid unused variable lint errors
	_ = holidaysClient
	_ = fundingClient
}
