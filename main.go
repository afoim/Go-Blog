package main

import (
    "bufio"
    "flag"
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

type Post struct {
    Title            string
    Content          template.HTML
    Date             time.Time
    Filename         string
    OriginalFilename string
}

type PageData struct {
    Posts []Post
    Today time.Time
    Post  Post
}

func anchorize(title string) string {
    return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func main() {
    // 命令行参数
    createPost := flag.String("c", "", "创建新文章，指定md文件名，并且作为文章标题")
    flag.Parse()

    // 如果提供了创建文章的标题，则执行创建文章操作
    if *createPost != "" {
        err := createNewPost(*createPost)
        if err != nil {
            log.Fatal("创建新文章失败:", err)
        }
        return
    }

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

func createNewPost(title string) error {
    // 格式化当前日期
    dateStr := time.Now().Format("2006-01-02")
    // 格式化文件名，去掉空格并转换为小写
    filename := strings.ToLower(strings.ReplaceAll(title, " ", "-")) + ".md"

    // 创建新文章内容，仅包含 !url: 和 !create:
    content := fmt.Sprintf("!url: \n!create: %s\n", dateStr)

    // 写入文件
    err := ioutil.WriteFile(filepath.Join("posts", filename), []byte(content), 0644)
    if err != nil {
        return err
    }

    fmt.Printf("新文章已创建: %s\n", filename)
    return nil
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
                createDate = time.Time{}
            }
        } else {
            if lineCount > 2 {
                remainingContent.WriteString(line + "\n")
            }
        }
    }

    return shortLink, createDate, remainingContent.String()
}

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

            shortLinkCode, createDate, remainingContent := parsePostInfo(string(content))

            extensions := parser.CommonExtensions | parser.AutoHeadingIDs
            p := parser.NewWithExtensions(extensions)

            opts := html.RendererOptions{
                Flags: html.CommonFlags | html.HrefTargetBlank,
            }
            renderer := html.NewRenderer(opts)

            htmlContent := markdown.ToHTML([]byte(remainingContent), p, renderer)

            originalFilename := strings.TrimSuffix(file.Name(), ".md")
            filename := originalFilename
            if shortLinkCode != "" {
                filename = shortLinkCode
            }

            htmlStr := string(htmlContent)

            for _, heading := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
                htmlStr = strings.ReplaceAll(htmlStr,
                    fmt.Sprintf("<%s>", heading),
                    fmt.Sprintf("<%s id=\"%s\">", heading, anchorize(originalFilename)))
            }

            post := Post{
                Title:            originalFilename,
                Content:          template.HTML(htmlStr),
                Date:             createDate,
                Filename:         filename,
                OriginalFilename: originalFilename,
            }

            posts = append(posts, post)
        }
    }

    sort.Slice(posts, func(i, j int) bool {
        return posts[i].Date.After(posts[j].Date)
    })

    return posts, nil
}

func generatePages(posts []Post) error {
    tmpl, err := template.ParseFiles("templates.html")
    if err != nil {
        return err
    }

    indexFile, err := os.Create("dist/index.html")
    if err != nil {
        return err
    }
    defer indexFile.Close()

    err = tmpl.Execute(indexFile, PageData{
        Posts: posts,
        Today: posts[0].Date,
    })
    if err != nil {
        return err
    }

    for _, post := range posts {
        file, err := os.Create(filepath.Join("dist", post.Filename+".html"))
        if err != nil {
            return err
        }

        err = tmpl.Execute(file, PageData{
            Post:  post,
            Today: post.Date,
        })
        file.Close()
        if err != nil {
            return err
        }
    }

    return nil
}
