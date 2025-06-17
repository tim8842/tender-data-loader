package reader

import "os"

func ReadHtmlFile(path string) []byte {
	if data, err := os.ReadFile(path); err == nil {
		return data
	}
	return nil
}
