package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"sort"
	"strings"
	"time"
)

// TestResult 存储单次测试的结果
type TestResult struct {
	StatusCode    int
	StatusText    string
	ResponseSize  int64
	DNSLookup     time.Duration
	TCPConnection time.Duration
	TLSHandshake  time.Duration
	TTFB          time.Duration
	DownloadTime  time.Duration
	TotalTime     time.Duration
	Error         error
}

// TestStatistics 存储多次测试的统计信息
type TestStatistics struct {
	Rounds  int
	Fastest time.Duration
	Slowest time.Duration
	Average time.Duration
	Success int
	Failed  int
}

// TestWebsite 测试网站访问情况
func TestWebsite(url string, rounds int) {
	if rounds <= 0 {
		rounds = 3 // 默认3轮
	}

	fmt.Printf("Testing: %s\n", url)
	fmt.Printf("Rounds: %d\n", rounds)
	fmt.Println()

	// 确保URL格式正确
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	results := make([]TestResult, 0, rounds)

	for i := 0; i < rounds; i++ {
		result := testSingleRound(url)
		results = append(results, result)
	}

	// 统计结果
	stats := calculateStatistics(results)

	// 显示最近一次测试的详细信息
	lastResult := results[len(results)-1]
	if lastResult.Error == nil {
		printDetailedResult(lastResult)
	} else {
		printErrorResult(lastResult)
	}

	// 显示统计信息
	printStatistics(stats)

	// 显示等级
	printGrade(stats)
}

// testSingleRound 执行单次测试
func testSingleRound(url string) TestResult {
	result := TestResult{}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Errorf("创建请求失败: %w", err)
		return result
	}

	// 追踪请求各阶段的时间
	var (
		start        time.Time
		dnsStart     time.Time
		dnsDone      time.Time
		tcpStart     time.Time
		tcpDone      time.Time
		tlsStart     time.Time
		tlsDone      time.Time
		tfbDone      time.Time
		downloadDone time.Time
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(di httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
		},
		ConnectStart: func(network, addr string) {
			if tcpStart.IsZero() {
				tcpStart = time.Now()
			}
		},
		ConnectDone: func(network, addr string, err error) {
			tcpDone = time.Now()
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			tlsDone = time.Now()
		},
		GotFirstResponseByte: func() {
			tfbDone = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// 发送请求
	start = time.Now()
	resp, err := client.Do(req)
	downloadDone = time.Now()

	if err != nil {
		result.Error = fmt.Errorf("请求失败: %w", err)
		return result
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("读取响应失败: %w", err)
		return result
	}

	// 计算各阶段时间
	result.StatusCode = resp.StatusCode
	result.StatusText = resp.Status
	result.ResponseSize = int64(len(body))
	result.DNSLookup = dnsDone.Sub(dnsStart)
	result.TCPConnection = tcpDone.Sub(tcpStart)

	// 如果使用HTTPS，计算TLS握手时间
	if strings.HasPrefix(url, "https://") {
		result.TLSHandshake = tlsDone.Sub(tlsStart)
	}

	result.TTFB = tfbDone.Sub(start)
	result.DownloadTime = downloadDone.Sub(tfbDone)
	result.TotalTime = downloadDone.Sub(start)

	return result
}

// calculateStatistics 计算统计信息
func calculateStatistics(results []TestResult) TestStatistics {
	stats := TestStatistics{
		Rounds:  len(results),
		Success: 0,
		Failed:  0,
	}

	totalTime := time.Duration(0)
	successfulTimes := make([]time.Duration, 0)

	for _, result := range results {
		if result.Error == nil {
			stats.Success++
			totalTime += result.TotalTime
			successfulTimes = append(successfulTimes, result.TotalTime)
		} else {
			stats.Failed++
		}
	}

	if len(successfulTimes) > 0 {
		sort.Slice(successfulTimes, func(i, j int) bool {
			return successfulTimes[i] < successfulTimes[j]
		})

		stats.Fastest = successfulTimes[0]
		stats.Slowest = successfulTimes[len(successfulTimes)-1]
		stats.Average = totalTime / time.Duration(len(successfulTimes))
	}

	return stats
}

// printDetailedResult 打印详细测试结果
func printDetailedResult(result TestResult) {
	fmt.Println("┌────────────────────────────┬────────────┐")
	fmt.Println("│ Metric                     │ Result     │")
	fmt.Println("├────────────────────────────┼────────────┤")
	fmt.Printf("│ Status Code                │ %d %s │\n", result.StatusCode, strings.Split(result.StatusText, " ")[1])
	fmt.Printf("│ Response Size              │ %s      │\n", formatSize(result.ResponseSize))
	fmt.Printf("│ DNS Lookup                 │ %s      │\n", formatDuration(result.DNSLookup))
	fmt.Printf("│ TCP Connection             │ %s      │\n", formatDuration(result.TCPConnection))

	if result.TLSHandshake > 0 {
		fmt.Printf("│ TLS Handshake              │ %s      │\n", formatDuration(result.TLSHandshake))
	} else {
		fmt.Println("│ TLS Handshake              │ N/A        │")
	}

	fmt.Printf("│ Time To First Byte (TTFB)  │ %s      │\n", formatDuration(result.TTFB))
	fmt.Printf("│ Download Time              │ %s      │\n", formatDuration(result.DownloadTime))
	fmt.Printf("│ Total Time                 │ %s      │\n", formatDuration(result.TotalTime))
	fmt.Println("└────────────────────────────┴────────────┘")
	fmt.Println()
}

// printErrorResult 打印错误结果
func printErrorResult(result TestResult) {
	fmt.Println("┌────────────────────────────┬────────────┐")
	fmt.Println("│ Metric                     │ Result     │")
	fmt.Println("├────────────────────────────┼────────────┤")
	fmt.Printf("│ Status                     │ Error      │\n")
	fmt.Printf("│ Error Message              │ %s │\n", truncateString(result.Error.Error(), 14))
	fmt.Println("└────────────────────────────┴────────────┘")
	fmt.Println()
}

// printStatistics 打印统计信息
func printStatistics(stats TestStatistics) {
	fmt.Println("Statistics", fmt.Sprintf("(%d runs)", stats.Rounds))

	if stats.Success > 0 {
		fmt.Printf("  Fastest: %s\n", formatDuration(stats.Fastest))
		fmt.Printf("  Slowest: %s\n", formatDuration(stats.Slowest))
		fmt.Printf("  Average: %s\n", formatDuration(stats.Average))
	} else {
		fmt.Println("  No successful runs")
	}

	if stats.Failed > 0 {
		fmt.Printf("  Failed: %d\n", stats.Failed)
	}

	fmt.Println()
}

// printGrade 打印等级
func printGrade(stats TestStatistics) {
	if stats.Success == 0 {
		fmt.Println("Grade: F (All tests failed)")
		return
	}

	// 根据平均时间计算等级
	avgTime := stats.Average.Milliseconds()
	var grade string
	var description string

	switch {
	case avgTime < 100:
		grade = "A"
		description = "Excellent"
	case avgTime < 200:
		grade = "B"
		description = "Good"
	case avgTime < 500:
		grade = "C"
		description = "Fair"
	case avgTime < 1000:
		grade = "D"
		description = "Poor"
	default:
		grade = "F"
		description = "Very Poor"
	}

	fmt.Printf("Grade: %s (%s)\n", grade, description)
}

// formatSize 格式化文件大小
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.1f B", float64(bytes))
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
