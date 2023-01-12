package internal

import (
	"flag"
	"fmt"
	"strings"
)

func PrintHelp() {
	fmt.Println()

	fmt.Println(shColor("yellow:bold", "gotestiful"), "  ", shColor("gray", "gotest + beautiful"), "  ", "nicer go tests for everyone")
	fmt.Println(shColor("whitesmoke", "created by https://github.com/alex-parra"))

	fmt.Println()
	fmt.Println(shColor("gray", strings.Repeat("-", 60)))
	fmt.Println()

	chev := shColor("gray", "❯")
	fmt.Println(shColor("white:bold", "Examples:"))
	fmt.Println(chev, shColor("white", "gotestiful"), shColor("gray", "runs 'go test ./...'"))
	fmt.Println(chev, shColor("white", "gotestiful -cache=false"), shColor("gray", "runs 'go test -count=1 ./...'"))
	fmt.Println(chev, shColor("white", "gotestiful -v"), shColor("gray", "runs 'go test -v ./...'"))
	fmt.Println(chev, shColor("white", "gotestiful some/package"), shColor("gray", "runs 'go test some/package'"))
	fmt.Println(chev, shColor("white", "gotestiful init"), shColor("gray", "creates default config at ./.gotestiful"))

	fmt.Println()
	fmt.Println(shColor("gray", strings.Repeat("-", 60)))
	fmt.Println()

	fmt.Println(shColor("white:bold", "Configuration:"))
	fmt.Println("  Run 'gotestiful init' to create a default config file for you project")
	fmt.Println("  Use the 'exclude' key to specify package prefixes to ignore those packages from tests and coverage")

	fmt.Println()
	fmt.Println(shColor("gray", strings.Repeat("-", 60)))
	fmt.Println()

	fmt.Println(shColor("white:bold", "Available flags:"))
	flag.PrintDefaults()
}
