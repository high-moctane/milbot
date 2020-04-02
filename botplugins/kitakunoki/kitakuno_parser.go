package kitakunoki

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// kitakunoURL は帰宅の木が載っているサイトです。
const kitakunoURL = "http://www.chiba-museum.jp/jyumoku2014/kensaku/namae.html"

func kitakunoList() ([]*kitakunoEntry, error) {
	r, err := kitakunoFetch()
	if err != nil {
		return nil, fmt.Errorf("kitakunoki list error: %w", err)
	}

	res, err := kitakunoParse(r)
	if err != nil {
		return nil, fmt.Errorf("kitakunoki list error: %w", err)
	}

	return res, nil
}

// kitakunoFetch は帰宅の木のソースを入手します。
func kitakunoFetch() (io.Reader, error) {
	// 帰宅の木のサイトを手に入れます。
	resp, err := http.Get(kitakunoURL)
	if err != nil {
		return nil, fmt.Errorf("kitakuno parse error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kitakuno parse error: %d %s", resp.StatusCode, resp.Status)
	}

	res := new(bytes.Buffer)
	trans := transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder())
	io.Copy(res, trans)
	return res, nil
}

// kitakunoParse は帰宅の木のウェブページをパーズします。
func kitakunoParse(r io.Reader) ([]*kitakunoEntry, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("kitakuno parse error: %w", err)
	}

	res := []*kitakunoEntry{}
	errCh := make(chan error)

	doc.Find(".b0_na1 > a, .b0_na2 > a").Each(func(_ int, s *goquery.Selection) {
		name := s.Text()
		// "る" のエントリーはないので飛ばす
		if name == "－" {
			return
		}

		url, exists := s.Attr("href")
		if !exists {
			select {
			case errCh <- errors.New("no kitakunoki found"):
			default:
			}
		}

		res = append(res, &kitakunoEntry{name: name, url: absURL(url)})
	})

	select {
	case err := <-errCh:
		return nil, fmt.Errorf("kitakuno parse error: %w", err)
	default:
	}
	return res, nil
}

// absURL は絶対パスに変換します。
func absURL(relURL string) string {
	absURL := path.Join(path.Dir(kitakunoURL), relURL)
	return absURL[:5] + "/" + absURL[5:]
}

// kitakunoEntry は帰宅の木のエントリーです
type kitakunoEntry struct {
	name string
	url  string
}
