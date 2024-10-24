### 这是一个轻量博客生成器
#### 创建文章： `go run main.go -c "Hello World!"` ，你可以使用MarkDown语法编写文章。 `!url: xxx` 即文章的短链接
##### 然后使用 `go mod tidy && go run main.go` 构建，HTML会输出到 `dist` 目录
###### 在 `templates.html` 中可以自定义HTML