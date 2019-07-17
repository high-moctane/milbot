package verse

import (
	"regexp"
	"strings"
)

// blankRegexp には全角スペースが含まれるぞ注意(｀･ω･´)！！
var blankRegexp = regexp.MustCompile(`(\s|　)`)

// verseMorphs は句ごとに分かれた morpheme 列です
type verseMorphs []morphemes

// String は vm をいい感じのフォーマットの string にします
func (vm verseMorphs) String() string {
	strs := make([]string, 0, len(vm))
	for _, v := range vm {
		strs = append(strs, v.String())
	}
	return strings.Join(strs, " ／ ")
}

// find は text から format, jilitsugoFlags を満たした定型詩を見つけてきます
func find(text string, format []int, jiritsugoFlags []bool) []string {
	ans := []string{}

	input := blankRegexp.ReplaceAllString(text, "")
	morphs := parseText(input)
	verses := findAllVerses(morphs, format, jiritsugoFlags)
	for _, verse := range verses {
		versestr := verse.String()
		ans = append(ans, versestr)
	}

	return ans
}

// findAllVerses は morphs から format, jiritsugoFlags を満たす定型詩を見つけてきます
func findAllVerses(morphs morphemes, format []int, jiritsugoFlags []bool) []verseMorphs {
	ans := []verseMorphs{}

	for i := 0; i < len(morphs); i++ {
		verse, ok := topVerse(morphs[i:], format, jiritsugoFlags)
		if !ok {
			continue
		}
		ans = append(ans, verse)
	}

	return ans
}

// topVerse は morphs の頭が定型詩になっているかどうか判定します。
// 定型詩になっている場合は ok == bool で，その定型詩も一緒に返します。
// false の場合は見つかりませんでした。
func topVerse(morphs morphemes, format []int, jiritsugoFlags []bool) (verseMorphs, bool) {
	ans := verseMorphs{}
	head := 0 // topPhrase で探索開始する morphs のインデックス
	for i := 0; i < len(format); i++ {
		// morphs が尽きてしまったらおわり
		if head >= len(morphs) {
			return nil, false
		}

		phrase, ok := topPhrase(morphs[head:], format[i], jiritsugoFlags[i])
		if !ok {
			return nil, false
		}
		ans = append(ans, phrase)
		head += len(ans[len(ans)-1])
	}

	return ans, true
}

// topPhrase は morphs の先頭から morae の長さかつ jiritsugoFlag を満たしたフレーズをとってきます。
// とってこれなかったら ok == false です。
func topPhrase(morphs morphemes, morae int, jiritsugoFlag bool) (morphemes, bool) {
	// まず jiritsugoFlag についての場合分けをする
	if jiritsugoFlag && !morphs[0].isJiritsugo() {
		return nil, false
	}

	// つぎに morphs から morae を満たすように phrase に詰め込む
	phrase := morphemes{}
	sum := 0
	for _, m := range morphs {
		if m.isOmitPOS() {
			return nil, false
		}

		phrase = append(phrase, m)
		sum += m.moraLen

		if sum == morae {
			return phrase, true
		}
		if sum > morae {
			return nil, false
		}
	}

	return nil, false
}
