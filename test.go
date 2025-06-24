package main

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
	socks5 "github.com/txthinking/socks5"
)

func getViaSocks5(proxyAddr, username, password, targetURL string, timeout time.Duration) (string, error) {
	client, err := socks5.NewClient(proxyAddr, username, password, int(timeout.Seconds()), 0)
	if err != nil {
		return "", fmt.Errorf("NewClient: %w", err)
	}

	// Оборачиваем client.Dial в DialContext, совместимый с http.Transport
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, _ := net.SplitHostPort(addr)
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("DNS lookup failed for %s: %w", host, err)
		}
		realAddr := net.JoinHostPort(ips[0].String(), port)
		return client.Dial(network, realAddr)
	}

	transport := &http.Transport{
		DialContext: dialContext,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("NewRequest: %w", err)
	}
	req.Header.Set("User-Agent", "python-requests/2.32.2")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Do: %w", err)
	}
	defer resp.Body.Close()
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("ошибка распаковки gzip: %w", err)
		}
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body)) // требует импорт brotli
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("ReadAll: %w", err)
	}
	return string(b), nil
}

func main() {
	proxy := "pool.proxy.market:10899"
	user := "hWg3KXoX7tfj"
	pass := "RNW78Fm5"
	target := "https://zakupki.gov.ru/epz/contract/contractCard/common-info.html?reestrNumber=2041100854220000189"

	body, err := getViaSocks5(proxy, user, pass, target, 15*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Println(body[:500])
}
