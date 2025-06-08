package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	subCommands := []string{"serve", "push", "preview"}

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	servePort := serveCmd.Int("port", 23342, "Port to run the server on")
	serveHtmlDir := serveCmd.String("htmldir", "", "Directory for storing rendered HTML files (Default: fresh temporary directory)")

	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	pushServer := pushCmd.String("server", "http://localhost:23342", "Base URL of mdbin server to upload the file to")
	pushFile := pushCmd.String("file", "", "File to upload")

	previewCmd := flag.NewFlagSet("preview", flag.ExitOnError)
	previewPort := previewCmd.Int("port", 23344, "Port to run the local server on")
	previewFile := previewCmd.String("file", "", "Markdown file to monitor for changes")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: mdbin <command> [options]\n")
		fmt.Printf("Available commands: %s\n", subCommands)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])
		if serveCmd.Parsed() && serveCmd.NArg() > 0 {
			fmt.Printf("Unknown argument: \"%s\"\n", previewCmd.Arg(0))
			os.Exit(1)
		}
		CmdServe(*servePort, *serveHtmlDir)
	case "push":
		pushCmd.Parse(os.Args[2:])
		if pushCmd.Parsed() && pushCmd.NArg() > 0 {
			fmt.Printf("Unknown argument: \"%s\"\n", previewCmd.Arg(0))
			os.Exit(1)
		}
		CmdPush(*pushServer, *pushFile)
	case "preview":
		previewCmd.Parse(os.Args[2:])
		if previewCmd.Parsed() && previewCmd.NArg() > 0 {
			fmt.Printf("Unknown argument: \"%s\"\n", previewCmd.Arg(0))
			os.Exit(1)
		}
		CmdPreview(*previewPort, *previewFile)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Printf("Available commands: %s\n", subCommands)
		os.Exit(1)
	}
}
