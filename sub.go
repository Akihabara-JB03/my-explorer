package sub
import (
	"os"
	"path/filepath"
	"fmt"
)
func GetdirSize(root string) float64 {
	var total float64
	var count int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
            return nil
        }
    // ここに、1つのファイルやフォルダを見つけるたびにやりたいことを書く
		if d.Type()&(os.ModeSymlink|os.ModeIrregular) != 0 {
			return nil
		}
		if !d.IsDir() {

			info, err := d.Info()
			count++
            if count%1000 == 0 {
                fmt.Printf("\rScanning... %d files found\r\n", count)
            }
			if err != nil {
    			return nil // 失敗したら諦めて次へ（return err だと全部止まるぞ！）
			}
			total += float64(info.Size())
		} else {
			return nil
		}
    	return nil
	})
	return total
}