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
		Addr      string `json:"addr"`
		Country   string `json:"country"`
		Province  string `json:"province"`
		City      string `json:"city"`
		ISP       string `json:"isp"`
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
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

	fmt.Println("正在查询出口IP...")
	fmt.Println("----------------------------------")

	// 1. 查询B站API（国内）
	fmt.Println("api.live.bilibili.com(国内)")
	bilibiliIP, biliErr := getBilibiliIP(client)
	if biliErr == nil && bilibiliIP.Data.Addr != "" {
		fmt.Printf(" IP:%s\n", bilibiliIP.Data.Addr)
		fmt.Printf(" 位置:%s %s %s\n", bilibiliIP.Data.Country, bilibiliIP.Data.Province, bilibiliIP.Data.City)
	} else {
		fmt.Printf(" 错误: %v\n", biliErr)
	}
	fmt.Println()

	// 2. 查询腾讯API（国内）
	fmt.Println("i.news.qq.com(国内)")
	qqIP, qqErr := getQQIP(client)
	if qqErr == nil && qqIP.IP != "" {
		fmt.Printf(" IP:%s\n", qqIP.IP)
		fmt.Printf(" 位置:%s %s %s\n", qqIP.Country, qqIP.Province, qqIP.City)
	} else {
		fmt.Printf(" 错误: %v\n", qqErr)
	}
	fmt.Println()

	// 3. 查询Cloudflare API（国际）
	fmt.Println("Cloudflare(国际)")
	cfIP, cfErr := getCloudflareIP(client)
	if cfErr == nil && cfIP != "" {
		fmt.Printf(" IP:%s\n", cfIP)
		// 获取基本信息显示
		ipinfo, _ := getIpinfoIP(client, cfIP)
		if ipinfo != nil {
			fmt.Printf(" Location:%s %s %s\n", ipinfo.Country, ipinfo.City, ipinfo.Org)
		}
	} else {
		fmt.Printf(" 错误: %v\n", cfErr)
	}
	fmt.Println()

	// 4. 查询IPinfo.io API（国际）
	fmt.Println("IPinfo.io(国际)")
	// 先获取IP再查询详细信息
	cfIPForIpinfo, _ := getCloudflareIP(client) // 复用Cloudflare的IP
	if cfIPForIpinfo != "" {
		ipinfo, ipinfoErr := getIpinfoIP(client, cfIPForIpinfo)
		if ipinfoErr == nil {
			fmt.Printf(" IP:%s\n", ipinfo.IP)
			fmt.Printf(" Location:%s %s %s\n", ipinfo.Country, ipinfo.City, ipinfo.Org)
		} else {
			fmt.Printf(" 错误: %v\n", ipinfoErr)
		}
	} else {
		// 如果Cloudflare失败，直接尝试IPinfo
		// 注意：ipinfo.io直接访问会返回当前IP
		resp, err := client.Get("https://ipinfo.io/json")
		if err != nil {
			fmt.Printf(" 错误: %v\n", err)
		} else {
			defer resp.Body.Close()
			var result IpinfoResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				fmt.Printf(" 错误: %v\n", err)
			} else {
				fmt.Printf(" IP:%s\n", result.IP)
				fmt.Printf(" Location:%s %s %s\n", result.Country, result.City, result.Org)
			}
		}
	}
	fmt.Println("----------------------------------")
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
