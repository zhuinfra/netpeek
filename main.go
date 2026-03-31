package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/zhuinfra/netpeek/dns"
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
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "num-rounds",
						Aliases: []string{"n"},
						Usage:   "Number of test rounds",
						Value:   3,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() < 1 {
						return fmt.Errorf("Error: URL is required")
					}
					url := c.Args().First()
					rounds := c.Int("num-rounds")
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
				Name:        "dns",
				Aliases:     []string{"dns-check"},
				Usage:       "Check for DNS hijacking",
				Description: "Check for DNS hijacking by comparing results from multiple DNS servers (Local, Aliyun, Cloudflare, Google).",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "type",
						Aliases: []string{"t"},
						Usage:   "DNS record type (A for IPv4, AAAA for IPv6)",
						Value:   "A",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() < 1 {
						return fmt.Errorf("Error: Domain is required")
					}
					domain := c.Args().First()
					queryType := c.String("type")
					dns.CheckHijacking(domain, queryType)
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
