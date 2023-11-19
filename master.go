package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	failureWaitTime = 5 // 请求失败后等待的时间，单位：秒
	maxAttempts     = 10
)

func main() {
	// 打印架构和系统
	arch := runtime.GOARCH
	system := runtime.GOOS
	go_version := runtime.Version()

	fmt.Printf("Running on \033[0;30;43m%s-%s\033[0m system\n", arch, system)
	fmt.Printf("Go version: \033[0;30;44m%s\033[0m\n", go_version)

	// 打印可执行文件名
	fmt.Println("Reading the executable filename...")
	filename := getFilename()
	fmt.Printf("Executable filename: \033[0;30;47m%s\033[0m\n", filename)

	// 提取学号和密码
	id, passwd := extractIDAndPassword(filename)
	
	fmt.Printf("id: \033[0;37;45m%s\033[0m\n", id)
	fmt.Printf("passwd: \033[0;30;46m%s\033[0m\n", passwd)

	urlHTTP := fmt.Sprintf("http://10.0.254.125:801/eportal/portal/login?&user_account=%s&user_password=%s", id, passwd)
	// urlHTTPS := fmt.Sprintf("https://auth.cqnu.edu.cn:802/eportal/portal/login?&user_account=%s&user_password=%s", id, passwd)
	// fmt.Printf("request url: \033[0;37;44m%s\033[0m\n", urlHTTP)

	// Windows下等待物理连接建立
	if isWindows() {
		fmt.Println("Waiting for physical connection...")
		time.Sleep(7 * time.Second)
	}

	// 尝试请求10次，直到请求成功
	for i := 0; i < maxAttempts; i++ {
		response, err := http.Get(urlHTTP)
		if err == nil {
			fmt.Printf("request status code: \033[0;37;42m%d\033[0m\n", response.StatusCode)
			if response.StatusCode != 200 {
				fmt.Println("Request failed. Retrying...")
				time.Sleep(failureWaitTime * time.Second)
				continue
			} else {
				fmt.Println("Request succeeded. Congratulations!")
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				time.Sleep(failureWaitTime * time.Second)
				return
			}

			// 处理 JSON 数据
			msg := extractJSONData(string(body))
			fmt.Printf("json.msg: \033[0;37;41m%s\033[0m\n", msg)
			time.Sleep(2 * time.Second)
			return // 请求成功，退出程序
		} else {
			fmt.Printf("An error occurred during the request: %v\n", err)
			time.Sleep(failureWaitTime * time.Second)
		}
	}
	fmt.Println("Exceeded maximum number of attempts. Request failed.")
}

func getFilename() string {
	filename := filepath.Base(os.Args[0])
	if isWindows() {
		// Windows系统下去掉.exe后缀
		filename = strings.TrimSuffix(filename, ".exe")
	}
	return filename
}

func extractIDAndPassword(filename string) (string, string) {
	parts := strings.Split(filename, ";")
	if len(parts) != 2 {
		fmt.Println("\033[0;37;41m Please make sure your id and password are separated by ';'. \033[0m")
		time.Sleep(failureWaitTime * time.Second)
		os.Exit(1)
	}

	id, passwd := parts[0], parts[1]
	if len(id) != 13 || !isNumeric(id) {
		fmt.Println("ID must be a 13-digit number.")
		time.Sleep(failureWaitTime * time.Second)
		os.Exit(1)
	}

	return id, passwd
}

func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isWindows() bool {
	// 简化处理，假定Windows系统的可执行文件后缀是".exe"
	return strings.HasSuffix(filepath.Base(os.Args[0]), ".exe")
}

func extractJSONData(responseText string) string {
	startIndex := strings.Index(responseText, `"msg":"`)
	startIndex += len(`"msg":"`)
	endIndex := strings.Index(responseText[startIndex:], `"`)
	return responseText[startIndex : startIndex+endIndex]
}
