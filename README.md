# go-sitemap

生成 `sitemap.xml` 的小工具。两种用法：直接给一串 URL，或者丢一个本地 HTML 文件让它把里面的链接全爬出来。

零依赖，只用 Go 标准库。

## 用法

```
# 直接列出 URL
go-sitemap https://example.com/ https://example.com/about

# 从 HTML 文件爬链接，相对路径按 -base 拼成绝对地址
go-sitemap -f index.html -base https://example.com

# 输出到文件而不是打印
go-sitemap -o sitemap.xml https://example.com/
```

选项：
- `-f <文件>`：从本地 HTML 文件提取 `<a href>` 链接
- `-base <地址>`：配合 `-f`，把相对链接拼成绝对 URL
- `-o <文件>`：写到文件，不写则打印到标准输出

## 说明

- 结果是排序后去重的，同一个 URL 只出现一次
- 锚点（`#`）、`javascript:`、`mailto:` 链接会被跳过
- 用正则提取链接，够应付大多数静态页面，不需要额外的 HTML 解析依赖
