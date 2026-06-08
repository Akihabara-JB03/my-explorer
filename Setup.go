package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	// Windowsのレジストリ（環境変数）を安全に書き換えるための外部ライブラリ
	"golang.org/x/sys/windows/registry"
)

// 【超重要】Readme.txt をここに書き足したから、もう0バイトの抜け殻にはならないぜ！
//go:embed my_explorer.exe license.txt Readme.txt
var content embed.FS

// 1F（青背景・白文字）を絶対に壊さないための色の定義
const (
	// プログレスバーの緑
	ColorGreen = "\033[32m"
	// 背景を青(44)、文字を白(37)に強制固定する命令（Resetの代わり）
	ColorFixed = "\033[37;44m"
)

func main() {
	// 1. 起動時に画面全体を塗りつぶす (Windowsのみ)
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "color 1F && cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	fmt.Println("======================================")
	fmt.Println("       my_explorer Installer          ")
	fmt.Println("======================================")

	// 2. ライセンス表示
	licenseData, err := content.ReadFile("license.txt")
	if err == nil {
		fmt.Println("\n【Software License Agreement】")
		fmt.Println(string(licenseData))
		fmt.Println("--------------------------------------")
		fmt.Print("この規約に同意しますか？ (y/n): ")

		var choice string
		fmt.Scanln(&choice)
		if strings.ToLower(choice) != "y" {
			fmt.Println("インストールを中止しました。")
			time.Sleep(2 * time.Second)
			return
		}
	}

	// 3. インストール先の準備 (ユーザーのホームに専用フォルダを作る)
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("エラー: ホームディレクトリが見つかりません: %v\n", err)
		return
	}
	installDir := filepath.Join(home, "my_explorer_app")
	os.MkdirAll(installDir, 0755)

	// コピーするファイルの一覧に Readme.txt もバッチリ登録！
	files := []string{"my_explorer.exe", "license.txt", "Readme.txt"}

	fmt.Println("\nインストールを開始します...")

	for _, fileName := range files {
		fmt.Printf("\nCopying: %s\n", fileName)

		for i := 0; i <= 20; i++ {
			percent := i * 5
			// ColorResetを使わず、ColorFixedで青背景を維持！
			bar := ColorGreen + strings.Repeat("■", i) + ColorFixed + strings.Repeat(" ", 20-i)
			fmt.Printf("\r[%s] %d%%", bar, percent)
			time.Sleep(95 * time.Millisecond)
		}
		fmt.Println() // 改行して次のファイルへ

		data, _ := content.ReadFile(fileName)
		os.WriteFile(filepath.Join(installDir, fileName), data, 0755)
	}

	// 4. 環境変数 PATH への自動追加 ＆ Windows全体への即時通知ギミック
	fmt.Println("\n--------------------------------------")
	fmt.Println("環境変数 (PATH) を設定中...")
	
	err = addPathAndNotify(installDir)
	if err != nil {
		fmt.Printf("⚠️ PATHの設定に失敗しました: %v\n", err)
	} else {
		fmt.Println("-> 環境変数 PATH の自動設定 ＆ システム通知に成功！")
	}
	fmt.Println("--------------------------------------")
	time.Sleep(1 * time.Second) // 成功ログを見せるために一瞬待つ

	// 5. 【大改造】画面をリフレッシュして、同じ青画面内にReadmeをドン！
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "cls").Run()
	}

	fmt.Println("======================================")
	fmt.Println(" [Success!] すべてのインストールが完了！ ")
	fmt.Println("======================================")

	// 内蔵された Readme.txt を読み込んでそのまま画面に出力
	readmeData, err := content.ReadFile("Readme.txt")
	if err == nil {
		fmt.Println("\n【 my_explorer 取扱説明書 (README) 】")
		fmt.Println(string(readmeData))
		fmt.Println("--------------------------------------")
	} else {
		fmt.Println("\nReadmeの読み込みに失敗しました。")
	}

	fmt.Println("再起動なしで、新しいコマンドプロンプトから 'my_explorer' で起動できます。")
	fmt.Println("\nEnterキーを押すとインストーラーを終了します...")
	
	var wait string
	fmt.Scanln(&wait)
}

// レジストリを操作してユーザー環境変数 PATH に追加し、OS全体に通知する関数
func addPathAndNotify(targetPath string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	oldPath, _, err := k.GetStringValue("PATH")
	if err != nil {
		oldPath = ""
	}

	if strings.Contains(oldPath, targetPath) {
		fmt.Println("-> PATH はすでに設定されています。")
		return nil
	}

	newPath := oldPath
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += targetPath

	err = k.SetStringValue("PATH", newPath)
	if err != nil {
		return err
	}

	// Win32 APIによる全体通知
	var hwndBroadcast uintptr = 0xffff
	var wmSettingchange uintptr = 0x001A
	
	envStr, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return err
	}
	
	modUser32 := syscall.NewLazyDLL("user32.dll")
	procSendMessage := modUser32.NewProc("SendMessageTimeoutW")
	
	procSendMessage.Call(
		hwndBroadcast,
		wmSettingchange,
		0,
		uintptr(unsafe.Pointer(envStr)),
		0x0002, // SMTO_ABORTIFHUNG
		5000,   // 5秒タイムアウト
		0,
	)

	return nil
}