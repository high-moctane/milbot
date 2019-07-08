package kitakunoki

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// 帰宅の木のリストが手に入る URL
var kitakunoURL = "http://www.chiba-museum.jp/jyumoku2014/kensaku/namae.html"

// fetchHTML で帰宅の HTML を fetch する
func fetchHTML() (io.Reader, error) {
	resp, err := http.Get(kitakunoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	trans := transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder())
	io.Copy(buf, trans)
	return buf, nil
}

// parseHTML で帰宅の HTML をパースする
func parseHTML(r io.Reader) ([]entry, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	entries, err := parseFunc(doc)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// parseFunc は帰宅の HTML をパースする際に使う
func parseFunc(node *html.Node) ([]entry, error) {
	if node == nil {
		return nil, fmt.Errorf("nil node")
	}

	entries := []entry{}

	// 木のリンクの class 属性は "b0_na1" or "b0_na2" なので，
	// これで木のリンクかどうかを判別
	classAttr := ""
	validClassAttrs := []string{"b0_na1", "b0_na2"}

	// リンクを入れとく
	href := ""

	stack := nodeStack{node}

	for len(stack) > 0 {
		n := stack.pop()
		if n == nil {
			continue
		}

		if n.Type == html.ErrorNode {
			return nil, fmt.Errorf("error node")
		}

		if n.Type == html.ElementNode {
			if n.Data == "td" {
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						classAttr = attr.Val
						href = ""
						break
					}
				}
			} else if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						href = attr.Val
						break
					}
				}
			}
		} else if n.Type == html.TextNode {
			// フラグばっかりだー(´; ω ;｀)
			// ここで帰宅のエントリーかどうか判別する
			valid := true
			if href == "" {
				valid = false
			}
			found := false
			for _, attr := range validClassAttrs {
				if classAttr == attr {
					found = true
				}
			}
			if !found {
				valid = false
			}

			// ここがメイン
			if valid {
				entries = append(entries, entry{
					name: n.Data,
					url:  makeAbsURL(href),
				})
			}

			// この処理はすごく大事
			classAttr = ""
			href = ""
		}

		stack.push(n.NextSibling)
		stack.push(n.FirstChild)
	}

	return entries, nil
}

// makeAbsPath で完全な url を生成する
func makeAbsURL(relURL string) string {
	absURL := path.Join(path.Dir(kitakunoURL), relURL)
	return absURL[:5] + "/" + absURL[5:]
}

// golang は stack も自分で実装するよー
type nodeStack []*html.Node

// push するよー
func (ns *nodeStack) push(n *html.Node) {
	*ns = append(*ns, n)
}

// pop するよー
func (ns *nodeStack) pop() *html.Node {
	if len(*ns) < 0 {
		panic("pop from empty stack")
	}
	last := (*ns)[len(*ns)-1]
	*ns = (*ns)[:len(*ns)-1]
	return last
}
