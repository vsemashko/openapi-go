package main

import (
	"context"
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
	holidaysClient, err := holidayssdk.NewInternalClient("https://holidays-server.staging.stashaway.sg")
	if err != nil {
		log.Fatalf("Failed to create holidays client: %v", err)
	}

	// Since we're using a dummy URL, we'll get errors when trying to call real endpoints
	// In a real application, you would use actual service URLs

	// Example 1: Check if a date is a holiday
	fmt.Println("\nExample 1: Check if a date is a holiday")
	today := time.Now().Format("2006-01-02")
	fmt.Printf("Checking if today (%s) is a holiday...\n", today)

	isHolidayResp, err := holidaysClient.HolidaysInternalControllerIsHoliday(context.Background(), holidayssdk.HolidaysInternalControllerIsHolidayParams{
		Date: today,
	})
	if err != nil {
		log.Fatalf("Failed to check holidays: %v", err)
	}

	fmt.Printf("Is today (%s) a holiday? %s", today, isHolidayResp)

	// Example 2: Get upcoming holidays (internal endpoint)
	fmt.Println("\nExample 2: Get upcoming holidays")
	fmt.Printf("Getting holidays from %s for the next 30 days...\n", today)

	upcomingHolidays, err := holidaysClient.HolidaysInternalControllerGetUpcomingHolidays(context.Background(), holidayssdk.HolidaysInternalControllerGetUpcomingHolidaysParams{
		FromDate:   today,
		DaysLookup: 30,
	})
	if err != nil {
		log.Fatalf("Failed to get upcoming holidays: %v", err)
	}

	// For the example, show mock data:
	fmt.Printf("Upcoming holidays: %s\n", upcomingHolidays)

	// Example 3: Initialize funding client
	fmt.Println("\n=== Funding Client Example ===")
	// Using the NewInternalClient function which doesn't require a security source
	fmt.Println("Initializing funding client with NewInternalClient (no security source needed)")
	fundingClient, err := fundingsdk.NewInternalClient("https://funding-server.staging.stashaway.sg")
	if err != nil {
		log.Fatalf("Failed to create funding client: %v", err)
	}

	// Example 3: Get deposit records for an account (internal endpoint)
	fmt.Println("\nExample 3: Get deposit records for an account")
	accountUUID := "948a53e6-858a-46e1-a212-93c4a40b87db"
	fmt.Printf("Getting deposit records for account %s...\n", accountUUID)

	depositRecords, err := fundingClient.DepositRecordsInternalControllerGetDepositRecordsByAccount(context.Background(), fundingsdk.DepositRecordsInternalControllerGetDepositRecordsByAccountParams{
		AccountUuid: accountUUID,
	})

	// For the example, show mock data:
	fmt.Printf("Deposit records for account %s: %s\n", accountUUID, depositRecords)
}
