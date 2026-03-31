package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zhuinfra/netpeek/ip"
	"github.com/zhuinfra/netpeek/websocket"
)

func main() {
	// 定义命令行参数
	versionFlag := flag.Bool("version", false, "Print version information")
	helpFlag := flag.Bool("help", false, "Print help information")

	// 解析命令行参数
	flag.Parse()

	// 处理版本参数
	if *versionFlag {
		fmt.Println("NetPeek v0.1.0")
		fmt.Println("Network monitoring command-line tool")
		os.Exit(0)
	}

	// 处理帮助参数
	if *helpFlag || flag.NArg() == 0 {
		printHelp()
		os.Exit(0)
	}

	// 处理命令
	command := flag.Arg(0)
	switch command {
	case "ip":
		ip.GetPublicIP()
	case "ping":
		fmt.Println("Ping command not implemented yet")
	case "speed":
		fmt.Println("Speed test command not implemented yet")
	case "scan":
		fmt.Println("Port scan command not implemented yet")
	case "ws":
		if flag.NArg() < 2 {
			fmt.Println("Usage: netpeek ws <websocket-url>")
			fmt.Println("Example: netpeek ws wss://echo.websocket.org")
			os.Exit(1)
		}
		wsURL := flag.Arg(1)
		websocket.TestWebSocket(wsURL)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("NetPeek - Network monitoring command-line tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  netpeek [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ip        Show public IP address and location")
	fmt.Println("  ping      Check network connectivity")
	fmt.Println("  speed     Test network speed")
	fmt.Println("  scan      Scan network ports")
	fmt.Println("  ws        Test WebSocket connection (supports ws:// and wss://)")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}
