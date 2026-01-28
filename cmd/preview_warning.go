package cmd

import (
	"fmt"
	"time"
)

func printPreviewWarning() {
	fmt.Println("\033[31mPreview: This command is in preview, not supported under SLAs, and may change or break without notice.\033[0m")
	time.Sleep(2 * time.Second)
}
