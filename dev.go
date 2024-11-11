package main

import (
    "fmt"
    "github.com/fsnotify/fsnotify"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"
)

var (
    regenerateMutex sync.RWMutex
)

func main() {
    // 确保dist目录存在
    if err := os.MkdirAll("dist", 0755); err != nil {
        log.Fatal("创建dist目录失败:", err)
    }

    // 初始构建
    if err := generateSite(); err != nil {
        log.Fatal("初始网站生成失败:", err)
    }

    // 创建文件监控
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal("创建文件监控失败:", err)
    }
    defer watcher.Close()

    // 监控posts目录和templates.html
    dirsToWatch := []string{"posts", "templates.html"}
    for _, dir := range dirsToWatch {
        if err := watcher.Add(dir); err != nil {
            log.Fatal("监控", dir, "失败:", err)
        }
    }

    // 防抖动计时器，避免文件变化太频繁导致过多重新生成
    var debounceTimer *time.Timer
    var debounceInterval = 500 * time.Millisecond

    // 启动监控goroutine
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
                    // 重置定时器
                    if debounceTimer != nil {
                        debounceTimer.Stop()
                    }
                    debounceTimer = time.AfterFunc(debounceInterval, func() {
                        log.Println("检测到文件变化，重新生成网站")
                        if err := generateSite(); err != nil {
                            log.Printf("重新生成网站失败: %v\n", err)
                        }
                    })
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("文件监控错误:", err)
            }
        }
    }()

    // 设置静态文件服务器，带有并发保护
    http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        regenerateMutex.RLock()
        defer regenerateMutex.RUnlock()

        path := r.URL.Path
        if path == "/" {
            path = "/index.html"
        }
        http.ServeFile(w, r, filepath.Join("dist", path))
    }))

    // 启动开发服务器
    port := ":7770"
    fmt.Printf("开发服务器启动在 http://localhost%s\n", port)
    fmt.Println("按 Ctrl+C 停止服务器")
    log.Fatal(http.ListenAndServe(port, nil))
}

// generateSite 调用主程序生成网站
func generateSite() error {
    regenerateMutex.Lock()
    defer regenerateMutex.Unlock()

    cmd := exec.Command("go", "run", "main.go")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("生成网站失败: %v\n输出: %s", err, output)
    }
    log.Println("网站重新生成完成")
    return nil
}