package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Analog wget application

func main() {
	url := "https://golang.org/dl/go1.16.3.linux-amd64.tar.gz"
	//extension := getExtension(url)
	err := download(url, ".tar.gz")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getExtension(url string) string {
	return url[strings.LastIndex(url, "."):]
}

func download(url string, extension string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	timeStamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("image%s%s", timeStamp, extension)

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, 1024)
	var receivedBytes int64
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		receivedBytes += int64(n)
		percent := float64(receivedBytes) / float64(resp.ContentLength) * 100

		//fmt.Print("\033[H\033[2J")
		fmt.Printf("Downloaded %d%%\n", int(percent))
	}

	fmt.Printf("Файл %s успешно загружен\n", fileName)
	return nil
}

func clearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
