package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/proxy/img", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" {
		http.Error(w, "Missing 'url' in query string", http.StatusBadRequest)
		return
	}

	imageURL = fixURL(imageURL)
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching image from: %s", parsedURL)

	// 创建HTTP客户端
	client := &http.Client{}

	// 使用解析后的URL创建请求
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 设置User-Agent头部
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0")

	// 发送请求并处理响应
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching image: %v", err)
		http.Error(w, "Failed to fetch image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Image not found", resp.StatusCode)
		return
	}

	// 设置响应头并转发内容
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(http.StatusOK)
	if _, err = io.Copy(w, resp.Body); err != nil {
		log.Printf("Error forwarding image data: %v", err)
	}
}

// 添加这个函数来检查并修正URL格式
func fixURL(urlStr string) string {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
	}
	return urlStr
}
