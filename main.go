package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// API响应结构体
type QQIPResponse struct {
	Ret      int    `json:"ret"`
	ErrMsg   string `json:"errMsg"`
	IP       string `json:"ip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
}

type BilibiliIPResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Addr      string  `json:"addr"`
		Country   string  `json:"country"`
		Province  string  `json:"province"`
		City      string  `json:"city"`
		ISP       string  `json:"isp"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"data"`
}

type CloudflareIPResponse struct {
	IP string `json:"ip"`
}

type IpinfoResponse struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

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
		getPublicIP()
	case "ping":
		fmt.Println("Ping command not implemented yet")
	case "speed":
		fmt.Println("Speed test command not implemented yet")
	case "scan":
		fmt.Println("Port scan command not implemented yet")
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
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// getPublicIP 获取本机出口IP并区分国内/海外
func getPublicIP() {
	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 先尝试国内API
	fmt.Println("正在查询出口IP...")

	// 尝试腾讯API
	qqIP, qqErr := getQQIP(client)
	if qqErr == nil && qqIP.IP != "" {
		fmt.Printf("出口IP: %s\n", qqIP.IP)
		fmt.Printf("位置: %s %s %s\n", qqIP.Country, qqIP.Province, qqIP.City)
		if qqIP.ISP != "" {
			fmt.Printf("运营商: %s\n", qqIP.ISP)
		}
		fmt.Println("网络类型: 国内网络")
		return
	}

	// 尝试B站API
	bilibiliIP, biliErr := getBilibiliIP(client)
	if biliErr == nil && bilibiliIP.Data.Addr != "" {
		fmt.Printf("出口IP: %s\n", bilibiliIP.Data.Addr)
		fmt.Printf("位置: %s %s %s\n", bilibiliIP.Data.Country, bilibiliIP.Data.Province, bilibiliIP.Data.City)
		if bilibiliIP.Data.ISP != "" {
			fmt.Printf("运营商: %s\n", bilibiliIP.Data.ISP)
		}
		fmt.Println("网络类型: 国内网络")
		return
	}

	// 国内API失败，尝试海外API
	fmt.Println("国内API查询失败，尝试海外API...")

	// 尝试Cloudflare API
	cfIP, cfErr := getCloudflareIP(client)
	if cfErr == nil && cfIP != "" {
		// 再用ipinfo.io获取详细信息
		ipinfo, ipinfoErr := getIpinfoIP(client, cfIP)
		if ipinfoErr == nil {
			fmt.Printf("出口IP: %s\n", ipinfo.IP)
			fmt.Printf("位置: %s %s %s\n", getCountryName(ipinfo.Country), ipinfo.Region, ipinfo.City)
			if ipinfo.Org != "" {
				fmt.Printf("运营商: %s\n", ipinfo.Org)
			}
			if ipinfo.Timezone != "" {
				fmt.Printf("时区: %s\n", ipinfo.Timezone)
			}
		} else {
			fmt.Printf("出口IP: %s\n", cfIP)
		}
		fmt.Println("网络类型: 海外网络")
		return
	}

	fmt.Println("所有API查询失败，请检查网络连接")
}

// getQQIP 从腾讯API获取IP信息
func getQQIP(client *http.Client) (*QQIPResponse, error) {
	resp, err := client.Get("https://i.news.qq.com/api/ip2city")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QQIPResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Ret != 0 {
		return nil, fmt.Errorf("API error: %s", result.ErrMsg)
	}

	return &result, nil
}

// getBilibiliIP 从B站API获取IP信息
func getBilibiliIP(client *http.Client) (*BilibiliIPResponse, error) {
	resp, err := client.Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getIpInfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BilibiliIPResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

// getCloudflareIP 从Cloudflare API获取IP信息
func getCloudflareIP(client *http.Client) (string, error) {
	resp, err := client.Get("https://www.cloudflare.com/cdn-cgi/trace")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析Cloudflare响应（文本格式）
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ip=") {
			return strings.TrimPrefix(line, "ip="), nil
		}
	}

	return "", fmt.Errorf("IP not found in response")
}

// getIpinfoIP 从ipinfo.io获取IP详细信息
func getIpinfoIP(client *http.Client, ip string) (*IpinfoResponse, error) {
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result IpinfoResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// getCountryName 将国家代码转换为中文名称
func getCountryName(code string) string {
	countryMap := map[string]string{
		"US": "美国",
		"HK": "中国香港",
		"CN": "中国",
		"JP": "日本",
		"KR": "韩国",
		"SG": "新加坡",
		"GB": "英国",
		"DE": "德国",
		"FR": "法国",
		"AU": "澳大利亚",
		"CA": "加拿大",
	}

	if name, ok := countryMap[code]; ok {
		return name
	}
	return code
}
