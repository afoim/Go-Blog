### 这是一个轻量博客生成器
#### 创建文章： `go run main.go -c "Hello World!"` ，你可以使用MarkDown语法编写文章。 `!url: xxx` 即文章的短链接
#### 本地预览： `go run dev.go` ，会启动一个本地服务器，访问 `http://localhost:7770` 即可预览
##### 构建&发布： `go mod tidy && go run main.go` ，构建完成后，HTML会输出到 `dist` 目录。可以将其托管到Cloudflare Pages、GitHub Pages、Netlify、Vercel等静态托管平台。
###### 在 `templates.html` 中可以自定义HTML

---

Demo：https://go-blog.acofork.us.kg
