package main

import (
    "fmt"
    "github.com/gomarkdown/markdown"
    "html/template"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// Post 结构体用于存储博文信息
type Post struct {
    Title    string
    Content  template.HTML
    Date     time.Time
    Filename string
}

// 模板字符串
const indexTemplate = `
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>我的博客</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 1rem;
        }
        .post-list {
            list-style: none;
            padding: 0;
        }
        .post-item {
            margin-bottom: 1rem;
            padding: 1rem;
            border-bottom: 1px solid #eee;
        }
        .post-date {
            color: #666;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <h1>我的博客</h1>
    <ul class="post-list">
    {{range .}}
        <li class="post-item">
            <h2><a href="{{.Filename}}.html">{{.Title}}</a></h2>
            <div class="post-date">{{.Date.Format "2006-01-02"}}</div>
        </li>
    {{end}}
    </ul>
</body>
</html>
`

const postTemplate = `
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 1rem;
        }
        .post-content {
            margin-top: 2rem;
        }
        .post-date {
            color: #666;
            font-size: 0.9rem;
        }
        .back-link {
            display: inline-block;
            margin-bottom: 2rem;
            color: #666;
            text-decoration: none;
        }
        .back-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <a href="index.html" class="back-link">← 返回首页</a>
    <article>
        <h1>{{.Title}}</h1>
        <div class="post-date">{{.Date.Format "2006-01-02"}}</div>
        <div class="post-content">
            {{.Content}}
        </div>
    </article>
</body>
</html>
`

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

    // 生成首页
    err = generateIndex(posts)
    if err != nil {
        log.Fatal("生成首页失败:", err)
    }

    // 生成博文页面
    err = generatePosts(posts)
    if err != nil {
        log.Fatal("生成博文页面失败:", err)
    }

    fmt.Println("网站生成完成！文件保存在 dist 目录中")
}

// 加载所有博文
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

            // 转换 Markdown 为 HTML
            html := markdown.ToHTML(content, nil, nil)
            
            // 使用文件名作为标题（去掉.md后缀）
            title := strings.TrimSuffix(file.Name(), ".md")
            
            post := Post{
                Title:    title,
                Content:  template.HTML(html),
                Date:     file.ModTime(),
                Filename: title,
            }
            
            posts = append(posts, post)
        }
    }

    return posts, nil
}

// 生成首页
func generateIndex(posts []Post) error {
    tmpl, err := template.New("index").Parse(indexTemplate)
    if err != nil {
        return err
    }

    file, err := os.Create("dist/index.html")
    if err != nil {
        return err
    }
    defer file.Close()

    return tmpl.Execute(file, posts)
}

// 生成博文页面
func generatePosts(posts []Post) error {
    tmpl, err := template.New("post").Parse(postTemplate)
    if err != nil {
        return err
    }

    for _, post := range posts {
        file, err := os.Create(filepath.Join("dist", post.Filename+".html"))
        if err != nil {
            return err
        }

        err = tmpl.Execute(file, post)
        file.Close()
        if err != nil {
            return err
        }
    }

    return nil
}