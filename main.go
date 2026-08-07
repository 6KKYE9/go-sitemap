// go-sitemap 生成 sitemap.xml。可以手写一串 URL，也可以给一个本地 HTML
// 文件让它把里面的 <a href> 全爬出来，最后按字母排序去重输出。
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
)

// urlSet 是 sitemap 的顶层结构，对应 <urlset>。
type urlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	Urls    []urlEntry
}

type urlEntry struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

// extractLinks 从一段 HTML 里抠出所有 <a href="..."> 的链接。
// 不用完整的 HTML 解析器，正则足够应付大多数静态页面，也省一个依赖。
func extractLinks(html string) []string {
	re := regexp.MustCompile(`(?i)<a[^>]+href\s*=\s*["']([^"']+)["']`)
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(html, -1) {
		href := strings.TrimSpace(m[1])
		if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
			continue
		}
		if !seen[href] {
			seen[href] = true
			out = append(out, href)
		}
	}
	return out
}

// normalize 把相对路径拼到 base 上，变成绝对 URL；已经是绝对的就原样返回。
func normalize(base, link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	b, err := url.Parse(base)
	if err != nil {
		return link
	}
	ref, err := url.Parse(link)
	if err != nil {
		return link
	}
	return b.ResolveReference(ref).String()
}

// buildSitemap 把 URL 列表排好序、去重，组装成可序列化的结构。
func buildSitemap(urls []string) urlSet {
	seen := map[string]bool{}
	var uniq []string
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		uniq = append(uniq, u)
	}
	sort.Strings(uniq)
	set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, u := range uniq {
		set.Urls = append(set.Urls, urlEntry{Loc: u})
	}
	return set
}

// marshalSitemap 序列化成带 XML 声明和结尾换行的字节，供输出和测试复用。
func marshalSitemap(set urlSet) ([]byte, error) {
	b, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return []byte(xml.Header + string(b) + "\n"), nil
}

func main() {
	fromFile := flag.String("f", "", "从本地 HTML 文件爬取链接，配合 -base 拼绝对地址")
	base := flag.String("base", "", "相对链接拼到这个基地址上，例如 https://example.com")
	out := flag.String("o", "", "输出到文件，不写则打印到标准输出")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法: go-sitemap [-f 文件 -base 地址] [-o 输出] [URL...]")
		fmt.Fprintln(os.Stderr, "  直接给 URL 列表，或 -f 从 HTML 爬链接")
	}
	flag.Parse()

	var urls []string
	if *fromFile != "" {
		data, err := os.ReadFile(*fromFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "读取失败:", err)
			os.Exit(1)
		}
		links := extractLinks(string(data))
		for _, l := range links {
			if *base != "" {
				urls = append(urls, normalize(*base, l))
			} else {
				urls = append(urls, l)
			}
		}
	}
	urls = append(urls, flag.Args()...)

	if len(urls) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	set := buildSitemap(urls)
	outBytes, err := marshalSitemap(set)
	if err != nil {
		fmt.Fprintln(os.Stderr, "生成失败:", err)
		os.Exit(1)
	}
	doc := outBytes
	if *out != "" {
		if err := os.WriteFile(*out, doc, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "写入失败:", err)
			os.Exit(1)
		}
		fmt.Printf("已写入 %d 个 URL 到 %s\n", len(set.Urls), *out)
	} else {
		os.Stdout.Write(doc)
	}
}
