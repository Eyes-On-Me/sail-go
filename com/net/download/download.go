package download

import (
	"io"
	"net/http"
	"os"
)

// 下载
func Download(url string, file string) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	io.Copy(f, r.Body)
	defer f.Close()
	return nil
}
