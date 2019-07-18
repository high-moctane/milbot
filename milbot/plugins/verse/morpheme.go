package verse

// "github.com/ikawaha/kagome/tokenizer" は dep ではなくて gopath にダウンロードしました

import (
	"strings"
	"unicode/utf8"

	"github.com/ikawaha/kagome/tokenizer"
)

// Morpheme は形態素を表します
type morpheme struct {
	surface      string // surface は形態素の文字列を表します
	partOfSpeech string // partOfSpeech は品詞を表します
	pronounce    string // pronounce は発音を表します（カタカナ）
	moraLen      int    // moraLen は発音に必要な音数を表します
}

func newMorpheme(token *tokenizer.Token) *morpheme {
	f := token.Features()
	m := &morpheme{
		surface:      token.Surface,
		partOfSpeech: f[0],
		pronounce:    f[len(f)-2],
		moraLen:      getMoraLen(f[len(f)-2]),
	}
	return m
}

// getMoraLen は pronounce から発音に必要な音数を数えます
func getMoraLen(pronounce string) int {
	cnt := 0
loop:
	for _, r := range pronounce {
		// 拗音は音にならない
		for _, youon := range []rune{'ャ', 'ィ', 'ュ', 'ェ', 'ョ'} {
			if r == youon {
				continue loop
			}
		}
		cnt++
	}
	return cnt
}

// parseText は text を morpheme の列に分解します
func parseText(text string) []*morpheme {
	t := tokenizer.New()
	// tokens := t.Analyze(text, tokenizer.Search)
	tokens := t.Tokenize(text)
	ans := make([]*morpheme, 0, len(tokens))

	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		m := newMorpheme(&token)
		ans = append(ans, m)
	}
	return ans
}

// isJiritsugo は m が自立語かどうか判定します
func (m *morpheme) isJiritsugo() bool {
	if m == nil {
		return false
	}

	for _, ji := range jiritsugo {
		if m.partOfSpeech == ji {
			return true
		}
	}
	return false
}

// isOmitPOS は omitPartOfSpeech に該当するかどうかを返します
func (m *morpheme) isOmitPOS() bool {
	for _, omit := range omitPartOfSpeech {
		if m.partOfSpeech == omit {
			return true
		}
	}
	return false
}

// hasASCII は m.surface に ASCII 文字があるかどうかを判別します
func (m *morpheme) hasASCII() bool {
	return utf8.RuneCountInString(m.surface) == len(m.surface)
}

// morphemes は morpheme 列です
type morphemes []*morpheme

// String は surface をつなげてもとの文字列に戻します
func (ms morphemes) String() string {
	if ms == nil {
		return ""
	}

	sli := make([]string, 0, len(ms))
	for _, m := range ms {
		sli = append(sli, m.surface)
	}
	return strings.Join(sli, "")
}
