package main

import (
	"flag"
	"fmt"
	"os"
)

func printAccountInfo(a *AccountInfo) {
	fmt.Printf("%-11s %d\n%-11s %s\n%-11s %s\n%-11s %s\n", "Account ID:", a.AccountID, "Name:", a.Name, "Email:", a.Email, "Username:", a.Username)
}

func accountDisplayName(a AccountInfo) string {
	if a.Name != "" {
		return a.Name
	}
	if a.Username != "" {
		return a.Username
	}
	if a.Email != "" {
		return a.Email
	}
	if a.AccountID != 0 {
		return fmt.Sprintf("%d", a.AccountID)
	}
	return ""
}

func handleLogin(client *GerritClient) {
	if len(flag.Args()) != 2 {
		fmt.Println("Usage: gerritui auth [password]")
		os.Exit(1)
	}

	password := flag.Arg(1)

	err := SavePassword(client.host, client.username, password)
	if err != nil {
		fmt.Printf("Error saving password to OS keyring: %s\n", err.Error())
		os.Exit(1)
	}

	client.password = password

	accountInfo, err := client.GetAccountInfo()
	if err != nil {
		fmt.Printf("Error logging in: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Logged in as %s\n", accountInfo.Name)
}

func handleLogout(client *GerritClient) {
	err := DeletePasswordFor(client.host, client.username)
	if err != nil {
		fmt.Printf("Error deleting password from OS keyring: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Logged out")
}

func handleGetMe(client *GerritClient) {
	accountInfo, err := client.GetAccountInfo()
	if err != nil {
		fmt.Printf("Error getting account info: %s\n", err.Error())
		os.Exit(1)
	}

	printAccountInfo(accountInfo)
}

func handleNuke() {
	err := DeleteAllPasswords()
	if err != nil {
		fmt.Printf("Error nuking passwords: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("All passwords removed from OS keyring")
}

const (
	ChangeIDField = "Change ID"
	SubjectField  = "Subject"
	OwnerField    = "Owner"
)

func handleChanges(client *GerritClient) {
	changes, err := client.GetChanges()
	if err != nil {
		fmt.Printf("Error retrieving changes: %s\n", err.Error())
		os.Exit(1)
	}

	longestChangeID := len(ChangeIDField)
	longestSubject := len(SubjectField)
	longestOwner := len(OwnerField)

	for _, change := range changes {
		longestChangeID = max(longestChangeID, len(change.ChangeID))
		longestSubject = max(longestSubject, len(change.Subject))
		longestOwner = max(longestOwner, len(accountDisplayName(change.Owner)))
	}

	fmt.Printf("%-*s %-*s %-*s\n", longestChangeID, ChangeIDField, longestSubject, SubjectField, longestOwner, OwnerField)

	for _, change := range changes {
		fmt.Printf("%-*s %-*s %-*s\n", longestChangeID, change.ChangeID, longestSubject, change.Subject, longestOwner, accountDisplayName(change.Owner))
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Subcommand required [inbox|changes|login|logout|nuke|me]")
		os.Exit(1)
	}

	switch cmd := flag.Arg(0); cmd {
	case "inbox":
		fmt.Println("Doing inbox things")
	case "changes":
		client := NewGerritClient()
		handleChanges(client)
	case "login":
		client := NewGerritClient()
		handleLogin(client)
	case "logout":
		client := NewGerritClient()
		handleLogout(client)
	case "nuke":
		handleNuke()
	case "me":
		client := NewGerritClient()
		handleGetMe(client)
	default:
		fmt.Printf("Unsupported subcommand: %s\n", cmd)
		os.Exit(1)
	}
}
