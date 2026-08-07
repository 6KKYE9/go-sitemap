package main

import (
	"strings"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	html := `
		<a href="/a">A</a>
		<a href="/b">B</a>
		<a href="/a">重复</a>
		<a href="#top">锚点跳过</a>
		<a href="javascript:void(0)">脚本跳过</a>
		<a href="mailto:a@b.com">邮件跳过</a>
		<a HREF='/c'>大写属性</a>
	`
	links := extractLinks(html)
	want := []string{"/a", "/b", "/c"}
	if len(links) != len(want) {
		t.Fatalf("应有 %d 个链接，实际 %d: %v", len(want), len(links), links)
	}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("第 %d 个应为 %q，实际 %q", i, w, links[i])
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		base, link, want string
	}{
		{"https://example.com", "/a", "https://example.com/a"},
		{"https://example.com/dir/", "page", "https://example.com/dir/page"},
		{"https://example.com", "https://other.com/x", "https://other.com/x"},
	}
	for _, c := range cases {
		if got := normalize(c.base, c.link); got != c.want {
			t.Errorf("normalize(%q,%q)=%q，应为 %q", c.base, c.link, got, c.want)
		}
	}
}

func TestBuildSitemapSortDedup(t *testing.T) {
	set := buildSitemap([]string{"https://e.com/b", "https://e.com/a", "https://e.com/a", ""})
	if len(set.Urls) != 2 {
		t.Fatalf("应去重成 2 个，实际 %d", len(set.Urls))
	}
	if set.Urls[0].Loc != "https://e.com/a" || set.Urls[1].Loc != "https://e.com/b" {
		t.Errorf("排序去重结果不对: %v %v", set.Urls[0].Loc, set.Urls[1].Loc)
	}
	if set.Xmlns != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("命名空间缺失: %q", set.Xmlns)
	}
}

func TestSitemapXMLValid(t *testing.T) {
	set := buildSitemap([]string{"https://example.com/"})
	// 至少要能正常序列化成合法的 xml 头与 urlset 标签
	out, err := marshalSitemap(set)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `<urlset`) || !strings.Contains(s, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Errorf("生成的 XML 缺关键标签:\n%s", s)
	}
	if !strings.Contains(s, "<loc>https://example.com/</loc>") {
		t.Errorf("生成的 XML 缺少 loc:\n%s", s)
	}
}
