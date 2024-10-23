package main

import (
    "bufio"
    "fmt"
    "github.com/gomarkdown/markdown"
    "github.com/gomarkdown/markdown/html"
    "github.com/gomarkdown/markdown/parser"
    "html/template"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

// Post 结构体用于存储博文信息
type Post struct {
    Title            string
    Content          template.HTML
    Date             time.Time
    Filename         string // 用于生成HTML文件的名称（可能是短链接）
    OriginalFilename string // 原始md文件名（用于标题显示）
}

// PageData 结构体用于存储页面数据
type PageData struct {
    Posts []Post
    Today time.Time
    Post  Post // 用于单篇文章页面
}

// 自定义函数将标题转换为锚点 ID
func anchorize(title string) string {
    return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func main() {
    // 创建 dist 目录
    err := os.MkdirAll("dist", 0755)
    if err != nil {
        log.Fatal("创建 dist 目录失败:", err)
    }

    // 读取所有博文
    posts, err := loadPosts()
    if err != nil {
        log.Fatal("加载博文失败:", err)
    }

    // 生成所有页面
    err = generatePages(posts)
    if err != nil {
        log.Fatal("生成页面失败:", err)
    }

    fmt.Println("网站生成完成！文件保存在 dist 目录中")
}

func parsePostInfo(content string) (string, time.Time, string) {
    scanner := bufio.NewScanner(strings.NewReader(content))
    var createDate time.Time
    shortLink := ""
    lineCount := 0
    var remainingContent strings.Builder

    for scanner.Scan() {
        lineCount++
        line := scanner.Text()
        if lineCount == 1 && strings.HasPrefix(line, "!url:") {
            shortLink = strings.TrimSpace(strings.TrimPrefix(line, "!url:"))
        } else if lineCount == 2 && strings.HasPrefix(line, "!create:") {
            dateStr := strings.TrimSpace(strings.TrimPrefix(line, "!create:"))
            var err error
            createDate, err = time.Parse("2006-01-02", dateStr)
            if err != nil {
                createDate = time.Time{} // 返回零值时间
            }
        } else {
            // 将剩余的内容写入 remainingContent
            if lineCount > 2 {
                remainingContent.WriteString(line + "\n")
            }
        }
    }

    return shortLink, createDate, remainingContent.String()
}


// 修改后的 loadPosts 函数
func loadPosts() ([]Post, error) {
    var posts []Post

    files, err := ioutil.ReadDir("posts")
    if err != nil {
        return nil, err
    }

    for _, file := range files {
        if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
            content, err := ioutil.ReadFile(filepath.Join("posts", file.Name()))
            if err != nil {
                return nil, err
            }

            // 检查并解析短链接和创建日期
            shortLinkCode, createDate, remainingContent := parsePostInfo(string(content))

            // 创建 markdown 解析器
            extensions := parser.CommonExtensions | parser.AutoHeadingIDs
            p := parser.NewWithExtensions(extensions)

            // 创建 HTML 渲染器
            opts := html.RendererOptions{
                Flags: html.CommonFlags | html.HrefTargetBlank,
            }
            renderer := html.NewRenderer(opts)

            // 转换 Markdown 为 HTML
            htmlContent := markdown.ToHTML([]byte(remainingContent), p, renderer)

            // 使用文件名作为标题（去掉.md后缀）
            originalFilename := strings.TrimSuffix(file.Name(), ".md")

            // 确定最终的文件名（使用短链接或原始文件名）
            filename := originalFilename
            if shortLinkCode != "" {
                filename = shortLinkCode
            }

            // 将 HTML 内容转换为字符串
            htmlStr := string(htmlContent)

            // 替换标题以包含 ID
            for _, heading := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
                htmlStr = strings.ReplaceAll(htmlStr,
                    fmt.Sprintf("<%s>", heading),
                    fmt.Sprintf("<%s id=\"%s\">", heading, anchorize(originalFilename)))
            }

            post := Post{
                Title:            originalFilename,  // 标题使用原始文件名
                Content:          template.HTML(htmlStr),
                Date:             createDate,         // 使用解析后的创建日期
                Filename:         filename,            // 用于生成HTML文件的名称
                OriginalFilename: originalFilename,    // 保存原始文件名
            }

            posts = append(posts, post)
        }
    }

    // 按照日期降序排序
    sort.Slice(posts, func(i, j int) bool {
        return posts[i].Date.After(posts[j].Date)
    })

    return posts, nil
}

// 在生成首页时使用最新的日期
func generatePages(posts []Post) error {
    // 读取模板文件
    tmpl, err := template.ParseFiles("templates.html")
    if err != nil {
        return err
    }

    // 生成首页
    indexFile, err := os.Create("dist/index.html")
    if err != nil {
        return err
    }
    defer indexFile.Close()

    err = tmpl.Execute(indexFile, PageData{
        Posts: posts,
        Today: posts[0].Date, // 使用第一篇文章的创建日期
    })
    if err != nil {
        return err
    }

    // 生成每篇文章的页面
    for _, post := range posts {
        file, err := os.Create(filepath.Join("dist", post.Filename+".html"))
        if err != nil {
            return err
        }

        err = tmpl.Execute(file, PageData{
            Post:  post,
            Today: post.Date, // 单篇文章页面使用其创建日期
        })
        file.Close()
        if err != nil {
            return err
        }
    }

    return nil
}
