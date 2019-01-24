package http_client

import (
	"fmt"
)

func GetPath(port int, path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
}
