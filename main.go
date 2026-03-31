package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/zhuinfra/netpeek/http"
	"github.com/zhuinfra/netpeek/ip"
	"github.com/zhuinfra/netpeek/websocket"
)

func main() {
	// 创建CLI命令
	cmd := &cli.Command{
		Name:    "netpeek",
		Usage:   "Network monitoring command-line tool",
		Version: "v0.1.0",
		Description: "A powerful network monitoring tool with multiple features\n" +
			"Commands include IP lookup, HTTP testing, WebSocket testing, and more.",

		// 添加子命令
		Commands: []*cli.Command{
			{
				Name:    "ip",
				Aliases: []string{"i"},
				Usage:   "Show public IP address and location",
				Description: "Display your public IP address from multiple sources (both domestic and international).\n" +
					"Sources include: Bilibili API, Tencent API, Cloudflare API, and IPinfo.io API.",
				Action: func(ctx context.Context, c *cli.Command) error {
					ip.GetPublicIP()
					return nil
				},
			},
			{
				Name:    "http",
				Aliases: []string{"web", "h"},
				Usage:   "Test website access",
				Description: "Test website access performance with customizable rounds.\n" +
					"Provides detailed metrics including DNS lookup, TCP connection, TLS handshake, TTFB, and total time.",
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() < 1 {
						return fmt.Errorf("Error: URL is required")
					}
					url := c.Args().First()
					rounds := 3 // 默认3轮
					if c.NArg() >= 2 {
						roundsStr := c.Args().Get(1)
						// 转换为整数
						if roundsInt, err := strconv.Atoi(roundsStr); err == nil {
							rounds = roundsInt
						} else {
							return fmt.Errorf("Error: Rounds must be a valid number")
						}
					}
					http.TestWebsite(url, rounds)
					return nil
				},
			},
			{
				Name:    "ws",
				Aliases: []string{"websocket"},
				Usage:   "Test WebSocket connection",
				Description: "Test WebSocket connections with support for ws:// and wss:// protocols.\n" +
					"Establishes connection, sends test message, and measures response time.",
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() < 1 {
						return fmt.Errorf("Error: WebSocket URL is required")
					}
					wsURL := c.Args().First()
					websocket.TestWebSocket(wsURL)
					return nil
				},
			},
			{
				Name:        "ping",
				Aliases:     []string{"p"},
				Usage:       "Check network connectivity",
				Description: "Check network connectivity to specified hosts.",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("Ping command not implemented yet")
					return nil
				},
			},
			{
				Name:        "speed",
				Aliases:     []string{"s"},
				Usage:       "Test network speed",
				Description: "Test network upload and download speeds.",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("Speed test command not implemented yet")
					return nil
				},
			},
			{
				Name:        "scan",
				Aliases:     []string{"sc"},
				Usage:       "Scan network ports",
				Description: "Scan network ports on specified hosts.",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("Port scan command not implemented yet")
					return nil
				},
			},
		},
	}

	// 运行应用
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
