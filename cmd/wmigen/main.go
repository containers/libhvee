package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	r := regexp.MustCompile(`^\s*([a-zA-Z0-9_]+)\s+(REF)?\s*([a-zA-Z0-9_-]+)(\[\])?\s*([^;:]*=[^;]*)?;?$`)

	fmt.Println("Paste MWI/CIM model definition and type Ctrl-D")

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	fmt.Printf("\nTransformed:\n")
	fmt.Println("struct {")
	for _, line := range lines {
		match := r.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		val := match[1]
		switch strings.ToLower(val) {
		case "boolean":
			val = "bool"
		case "datetime":
			val = "time.Time"
		}

		var ref string
		if match[2] == "REF" {
			val = "string"
			ref = "REF to " + match[1]
		}

		suffix := match[5]
		if len(suffix) > 0 || len(ref) > 0 {
			suffix = " // " + ref + suffix
		}

		fmt.Printf("\t%s %s%s%s\n", match[3], match[4], val, suffix)
	}
	fmt.Println("}")
}
