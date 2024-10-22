package main

import (
    "fmt"
    "github.com/gomarkdown/markdown"
    "github.com/gomarkdown/markdown/html"
    "github.com/gomarkdown/markdown/parser"
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

// PageData 结构体用于存储页面数据
type PageData struct {
    Posts []Post
    Today time.Time
}

// 自定义函数将标题转换为锚点 ID
func anchorize(title string) string {
    return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

// 模板字符串
const commonStyles = `
<style>
    /* 代码块容器 */
    .code-block-wrapper {
        position: relative;
        margin: 1.5rem 0;
    }

    /* 代码块 */
    pre {
        position: relative;
        padding: 2rem 1rem 1rem 1rem;
        background: #282c34;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        max-height: 400px;
        overflow: hidden;
        transition: max-height 0.3s ease-out;
    }

    /* 展开的代码块 */
    pre.expanded {
        max-height: none;
    }

    /* 窗口装饰点 */
    pre::before {
        content: '';
        position: absolute;
        top: 12px;
        left: 12px;
        width: 12px;
        height: 12px;
        background: #ff5f56;
        border-radius: 50%;
        box-shadow: 18px 0 0 #ffbd2e, 36px 0 0 #27c93f;
    }

    /* 代码区域 */
    pre code {
        display: block;
        overflow-x: auto;
        padding: 0 !important;
        font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
        font-size: 0.9rem;
        line-height: 1.5;
        tab-size: 4;
    }

    /* 行内代码 */
    p code, li code {
        background-color: #f1f1f1;
        color: #e91e63;
        padding: 2px 4px;
        border-radius: 4px;
        font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
        font-size: 0.9em;
    }

    /* 折叠渐变效果 */
    .code-block-wrapper::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 0;
        width: 100%;
        height: 60px;
        background: linear-gradient(transparent, #282c34);
        pointer-events: none;
        opacity: 1;
        transition: opacity 0.3s ease;
    }

    .code-block-wrapper.expanded::after {
        opacity: 0;
    }

    /* 展开/收起按钮 */
    .expand-button {
        position: absolute;
        left: 50%;
        bottom: 10px;
        transform: translateX(-50%);
        background: #3a404b;
        border: none;
        color: #fff;
        padding: 4px 12px;
        border-radius: 4px;
        cursor: pointer;
        z-index: 10;
        font-size: 0.8rem;
        transition: background-color 0.2s;
    }

    .expand-button:hover {
        background: #4a5160;
    }

    /* 复制按钮 */
    .copy-button {
        position: absolute;
        right: 10px;
        top: 8px;
        background: #3a404b;
        border: none;
        color: #fff;
        padding: 4px 8px;
        border-radius: 4px;
        cursor: pointer;
        z-index: 10;
        font-size: 0.8rem;
        transition: background-color 0.2s;
    }

    .copy-button:hover {
        background: #4a5160;
    }

    .copy-button.copied {
        background: #27c93f;
    }
    .footer {
    margin-top: 2rem;
    padding: 1rem;
    text-align: center;
    font-size: 0.8rem;
    color: #666;
}

</style>
`

const commonScripts = `
<script>
    // 对代码块进行处理
    document.addEventListener('DOMContentLoaded', function() {
        const codeBlocks = document.querySelectorAll('pre code');
        
        codeBlocks.forEach((codeBlock, index) => {
            const wrapper = document.createElement('div');
            wrapper.className = 'code-block-wrapper';
            codeBlock.parentElement.parentNode.insertBefore(wrapper, codeBlock.parentElement);
            wrapper.appendChild(codeBlock.parentElement);

            // 创建复制按钮
            const copyButton = document.createElement('button');
            copyButton.className = 'copy-button';
            copyButton.textContent = '复制';
            copyButton.onclick = () => copyCode(codeBlock, copyButton);
            wrapper.appendChild(copyButton);

            // 如果代码超过400px（与CSS中的max-height相匹配），添加展开按钮
            if (codeBlock.parentElement.scrollHeight > 400) {
                const expandButton = document.createElement('button');
                expandButton.className = 'expand-button';
                expandButton.textContent = '展开';
                expandButton.onclick = () => toggleCode(wrapper, expandButton);
                wrapper.appendChild(expandButton);
            }
        });
    });

    // 复制代码功能
    function copyCode(codeBlock, button) {
        const text = codeBlock.textContent;
        navigator.clipboard.writeText(text).then(() => {
            button.textContent = '已复制';
            button.classList.add('copied');
            setTimeout(() => {
                button.textContent = '复制';
                button.classList.remove('copied');
            }, 2000);
        }).catch(err => {
            console.error('复制失败:', err);
            button.textContent = '复制失败';
            setTimeout(() => {
                button.textContent = '复制';
            }, 2000);
        });
    }

    // 展开/收起代码功能
    function toggleCode(wrapper, button) {
        const pre = wrapper.querySelector('pre');
        const isExpanded = wrapper.classList.contains('expanded');
        
        if (isExpanded) {
            wrapper.classList.remove('expanded');
            pre.classList.remove('expanded');
            button.textContent = '展开';
            // 滚动到代码块顶部
            wrapper.scrollIntoView({ behavior: 'smooth' });
        } else {
            wrapper.classList.add('expanded');
            pre.classList.add('expanded');
            button.textContent = '收起';
        }
    }
</script>
`
const indexTemplate = `
<!DOCTYPE html>
<html lang="zh">
<head>
    <link rel="icon" type="image/png" href="https://q2.qlogo.cn/headimg_dl?dst_uin=2973517380&spec=5">
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>二叉树树的博客</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/atom-one-dark.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/highlight.min.js"></script>
    <script>hljs.highlightAll();</script>
    ` + commonStyles + commonScripts + `
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 1rem;
            background-color: #fff;
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
    <h1>二叉树树的博客</h1>
    <ul class="post-list">
    {{range .Posts}}
        <li class="post-item">
            <h2><a href="{{.Filename}}.html">{{.Title}}</a></h2>
            <div class="post-date">{{.Date.Format "2006-01-02"}}</div>
        </li>
    {{end}}
    </ul>

    <div class="footer">
        <p>© {{.Today.Format "2006"}} 二叉树树 版权所有</p>
        <p>采用 <a href="https://creativecommons.org/licenses/by-nc-sa/4.0/">CC BY-NC-SA 4.0</a> 许可证</p>
    </div>
</body>
</html>
`

const postTemplate = `
<!DOCTYPE html>
<html lang="zh">
<head>
    <link rel="icon" type="image/png" href="https://q2.qlogo.cn/headimg_dl?dst_uin=2973517380&spec=5">
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Post.Title}}</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/styles/atom-one-dark.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.8.0/highlight.min.js"></script>
    <script>hljs.highlightAll();</script>
    ` + commonStyles + commonScripts + `
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 1rem;
            background-color: #fff;
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
        <h1>{{.Post.Title}}</h1>
        <div class="post-date">{{.Post.Date.Format "2006-01-02"}}</div>
        <div class="post-content">
            {{.Post.Content}}
        </div>
    </article>

    <div class="footer">
        <p>© {{.Today.Format "2006"}} 二叉树树 版权所有</p>
        <p>采用 <a href="https://creativecommons.org/licenses/by-nc-sa/4.0/">CC BY-NC-SA 4.0</a> 许可证</p>
    </div>
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

            // 创建 markdown 解析器
            extensions := parser.CommonExtensions | parser.AutoHeadingIDs
            p := parser.NewWithExtensions(extensions)

            // 创建 HTML 渲染器
            opts := html.RendererOptions{
                Flags: html.CommonFlags | html.HrefTargetBlank,
            }
            renderer := html.NewRenderer(opts)

            // 转换 Markdown 为 HTML
            htmlContent := markdown.ToHTML(content, p, renderer)

            // 使用文件名作为标题（去掉.md后缀）
            title := strings.TrimSuffix(file.Name(), ".md")

            // 将 HTML 内容转换为字符串
            htmlStr := string(htmlContent)

            // 替换标题以包含 ID
            for _, heading := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
                htmlStr = strings.ReplaceAll(htmlStr, fmt.Sprintf("<%s>", heading), fmt.Sprintf("<%s id=\"%s\">", heading, anchorize(title)))
            }

            post := Post{
                Title:    title,
                Content:  template.HTML(htmlStr),
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

    data := PageData{
        Posts: posts,
        Today: time.Now(),
    }

    return tmpl.Execute(file, data)
}

// PostPageData 结构体用于博文页面数据
type PostPageData struct {
    Post  Post
    Today time.Time
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

        data := PostPageData{
            Post:  post,
            Today: time.Now(),
        }

        err = tmpl.Execute(file, data)
        file.Close()
        if err != nil {
            return err
        }
    }

    return nil
}