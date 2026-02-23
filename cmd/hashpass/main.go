package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	} else {
		fmt.Print("Enter password to hash: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\nError reading password:", err)
			os.Exit(1)
		}
		password = string(bytePassword)
		fmt.Println() // Print a newline after password input
	}

	password = strings.TrimSpace(password)
	if password == "" {
		fmt.Println("Password cannot be empty")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Failed to hash password:", err)
		os.Exit(1)
	}

	fmt.Printf("\nBcrypt Hash:\n%s\n", string(hash))
	fmt.Println("\nYou can use this hash in your config.yml under auth.users.password_hash")
}
