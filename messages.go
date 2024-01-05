package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func printHeaderMessage() time.Time {
	startTime := time.Now()

	fmt.Printf("Script execution started at %s.\n", startTime.Format(time.UnixDate))

	return startTime
}

func printFooterMessage(s time.Time) {
	endTime := time.Now()

	fmt.Printf("Script execution ended at %s.\nTotal execution time is %s.\n", endTime.Format(time.UnixDate), endTime.Sub(s))
}

func printRecordAddedMessage(i int) {
	fmt.Printf("The record in line %d is updated successfully.\n", i+1)
}

func printFlagErrorMessageAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func printRegexErrorMessageAndExit(i int) {
	fmt.Printf("Google Drive link is invalid in line %d.\nFix the link, clear partially migrated records and re-run this script again.\n", i+1)
	os.Exit(1)
}
