package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"
)

// Analog wget application

const buffSize = 1024

func main() {
	startTime := time.Now().Format("23 Feb 2021 15:04:05")
	fmt.Println("Start time: ", startTime)

	url := "https://golang.org/dl/go1.16.3.linux-amd64.tar.gz"
	err := download(url)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func download(url string) error {
	fmt.Printf("Downloading from %s\n", url)
	fmt.Printf("Sending request, awaiting response... ")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	fmt.Printf("status: %d %s\n", status, http.StatusText(status))
	contType := resp.Header.Get("Content-Type")
	fmt.Printf("Content type: %s\n", contType)
	contLength := resp.ContentLength
	fmt.Printf("Content length: %dKb\n", contLength/1024.0)
	extension, err := mime.ExtensionsByType(contType)
	fmt.Printf("Saving to: /%s\n", getNameFromURL(url))

	name := getNameFromURL(url)
	fileName := fmt.Sprintf("%s%s", name, extension[0])

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var receivedBytes int64
	startTime := time.Now()

	speedLimit := 1.0
	buf := make([]byte, buffSize)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			fmt.Printf("\rФайл %s успешно загружен\n", fileName)
			break
		}
		receivedBytes += int64(n)
		percent := float64(receivedBytes) / float64(contLength) * 100
		speed := float64(receivedBytes) / time.Since(startTime).Seconds() / 1024.0 / 1024.0
		if speed > speedLimit {
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Printf("\rDownloaded %d%% %.2f Mb/s", int(percent), speed)
	}

	return nil
}

//
//func clearTerminal() {
//	var cmd *exec.Cmd
//	if runtime.GOOS == "windows" {
//		cmd = exec.Command("cmd", "/c", "cls")
//	} else {
//		cmd = exec.Command("clear")
//	}
//	cmd.Stdout = os.Stdout
//	if err := cmd.Run(); err != nil {
//		log.Fatal(err)
//	}
//}

func getNameFromURL(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}
