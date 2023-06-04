package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	err      error
	response *http.Response
	body     []byte
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("USAGE : %s <URL LIST FILE>\n", os.Args[0])
		os.Exit(0)
	}

	domain := os.Args[1]

	fileHandle, _ := os.Create("report.html")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	fmt.Fprintln(writer, "<!DOCTYPE html>")
	fmt.Fprintln(writer, "<html>")
	fmt.Fprintln(writer, "<style>")
	fmt.Fprintln(writer, "table, th, td {")
	fmt.Fprintln(writer, "  border:1px solid black;")
	fmt.Fprintln(writer, "}")
	fmt.Fprintln(writer, "</style>")
	fmt.Fprintln(writer, "<body>")
	fmt.Fprintln(writer, "<table style=\"width:100%\">")
	fmt.Fprintln(writer, "<tr>")
	fmt.Fprintln(writer, "<th>Proxy Target</th>")
	fmt.Fprintln(writer, "<th>Proxy Response Code</th>")
	fmt.Fprintln(writer, "<th>URL Attempt</th>")
	fmt.Fprintln(writer, "<th>Status</th>")
	fmt.Fprintln(writer, "<th>Length</th>")
	fmt.Fprintln(writer, "</tr>")

	f, err := os.Open(domain)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		targetURL := "https://google.com"

		for _, port := range []string{"80", "443", "8080", "8443"} {
			address := scanner.Text() + ":" + port
			conn, err := net.DialTimeout("tcp", address, 2*time.Second)
			if err != nil {
				fmt.Printf("Failed to connect to %s:%s: %s\n", scanner.Text(), port, err.Error())
				continue
			}

			fmt.Printf("Connected to %s:%s\n", scanner.Text(), port)
			conn.Close()

			proxyURL := "http://" + address
			proxy, err := url.Parse(proxyURL)
			if err != nil {
				log.Println(err)
			}

			transport := &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}

			client := &http.Client{
				Transport: transport,
			}

			req, err := http.NewRequest("GET", targetURL, nil)
			if err != nil {
				log.Println(err)
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/76.0")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Attempting proxy through:", proxyURL)
				log.Println(err)
				fmt.Fprintln(writer, "<tr>")
				fmt.Fprintln(writer, "<td>"+proxyURL+"</td>")
				fmt.Fprintln(writer, "<td>"+err.Error()+"</td>")
				fmt.Fprintln(writer, "<td><b><a href=\""+targetURL+"\" target=\"_blank\">"+targetURL+"</a></b></td>")
				fmt.Println("-------------------------------")
				fmt.Fprintln(writer, "<td></td>")
				fmt.Fprintln(writer, "<td></td>")
				fmt.Fprintln(writer, "</tr>")
			} else {
				fmt.Println("Proxy Response Code:", resp.StatusCode)
				length := strconv.Itoa(int(resp.ContentLength))
				status := strconv.Itoa(resp.StatusCode)
				fmt.Fprintln(writer, "<tr>")
				fmt.Fprintln(writer, "<td>"+proxyURL+"</td>")
				fmt.Fprintln(writer, "<td>"+status+"</td>")
				fmt.Fprintln(writer, "<td><b><a href=\""+targetURL+"\" target=\"_blank\">"+targetURL+"</a></b></td>")
				fmt.Println("-------------------------------")
				fmt.Fprintln(writer, "<td><b>"+status+"</b></td>")
				fmt.Fprintln(writer, "<td><b>"+length+"</b></td>")
				fmt.Fprintln(writer, "</tr>")
			}

			writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(writer, "</table>")
	fmt.Fprintln(writer, "</body>")
	fmt.Fprintln(writer, "</html>")
}
