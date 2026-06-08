package main
import (
	"fmt"
	"os"
	"path/filepath"
	"my_exprorer/sub"
)
func main() {
	
	if len(os.Args) < 2 {
		fmt.Println("パスを指定してください")
		return
	}

	readDir := os.Args[1]

	// パスが存在するか一応チェック
	if _, err := os.Stat(readDir); os.Args[1] == "" || err != nil {
		fmt.Println("有効なパスではありません")
		return
	}
	entries,err := os.ReadDir(readDir)
	if err != nil {
		fmt.Print("エラー\r\n")
		return
	}
	for _, entry := range entries {
		
		fmt.Printf("Scanning: %s ... ", entry.Name())
		fullPath := filepath.Join(readDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			fmt.Print("エラー\r\n")
			return
		}
		size :=float64(info.Size())
		if entry.IsDir() {
			total := sub.GetdirSize(fullPath)
			mib := total / 1048576
			if total >= 1024*1024 {
    			// 1MB以上ならMiBで表示
    			fmt.Printf("[dir] %s | Size:%.2fMiB\r\n", entry.Name(), mib)
			} else if total >= 1024 {
    			// 1KB以上1MB未満ならKiBで表示
				kib := total /1024
				fmt.Printf("[dir] %s | Size:%.2fKiB\r\n", entry.Name(), kib)
			} else {
				fmt.Printf("[dir] %s | Size:%.2fB\r\n", entry.Name(), total)
			}
		} else {
				if size >= 1024*1024 {
					mib := size / 1048576
					fmt.Printf("[File] %s | Size:%.2fMiB\r\n",entry.Name(),mib)
				} else if size >= 1024 {
					kib := size /1024
					fmt.Printf("[File] %s | Size:%.2fKiB\r\n", entry.Name(), kib)
				} else {
					fmt.Printf("[File] %s | Size:%.2fB\r\n", entry.Name(), size)
				}
		}
	}

}