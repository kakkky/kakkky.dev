package domain

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Slug string

const (
	SlugMaxLength            = 20
	derivedSlugFallbackChars = 8
)

var slugRegexp = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func NewSlug(slug string) (Slug, error) {
	if slug == "" {
		return "", ErrInvalidArgument.With("slug は 必須 です")
	}
	if len(slug) > SlugMaxLength {
		return "", ErrInvalidArgument.With(fmt.Sprintf("slug は %d 文字以内 です", SlugMaxLength))
	}
	if !slugRegexp.MatchString(slug) {
		return "", ErrInvalidArgument.With("slug は 英小文字、数字、ハイフン(-) のみ です")
	}

	return Slug(slug), nil
}

// DeriveSlug は title から article slug を機械的に導出する。
// 順序:
//  1. title を kagome で形態素解析し、各 token を「ASCII 素通し / 日本語は
//     読み (カタカナ) をローマ字化」して連結
//  2. 空になった場合 (kagome が読みを返さない記号のみ等) は
//     8 文字のランダム [a-z0-9] にフォールバック
//
// 戻り値は必ず NewSlug の validation を通る形式。
func DeriveSlug(title string) Slug {
	s := tokenizeToSlug(title)
	if s == "" {
		return Slug(randomSlug(derivedSlugFallbackChars))
	}
	return Slug(s)
}

var (
	kagomeOnce sync.Once
	kagomeTk   *tokenizer.Tokenizer
)

func kagomeTokenizer() *tokenizer.Tokenizer {
	kagomeOnce.Do(func() {
		t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
		if err != nil {
			panic("kagome init: " + err.Error())
		}
		kagomeTk = t
	})
	return kagomeTk
}

func tokenizeToSlug(title string) string {
	tokens := kagomeTokenizer().Tokenize(title)

	var parts []string
	for _, tk := range tokens {
		if tk.Class == tokenizer.DUMMY {
			continue
		}
		surface := tk.Surface
		if surface == "" {
			continue
		}

		// ASCII だけの surface は素通し (小文字化 / 記号除去)
		if isASCII(surface) {
			if p := normalizeASCIIToken(surface); p != "" {
				parts = append(parts, p)
			}
			continue
		}

		// 日本語含み: kagome の 読み (カタカナ) をローマ字化
		reading, ok := tk.Reading()
		if !ok {
			continue
		}
		if p := katakanaToRomaji(reading); p != "" {
			parts = append(parts, p)
		}
	}

	joined := strings.Join(parts, "-")
	joined = compactHyphens(joined)
	joined = strings.Trim(joined, "-")

	if len(joined) > SlugMaxLength {
		joined = joined[:SlugMaxLength]
		joined = strings.TrimRight(joined, "-")
	}
	return joined
}

func isASCII(s string) bool {
	for _, r := range s {
		if r >= 128 {
			return false
		}
	}
	return true
}

func normalizeASCIIToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '_', r == '-':
			b.WriteRune('-')
		}
	}
	return compactHyphens(strings.Trim(b.String(), "-"))
}

func compactHyphens(s string) string {
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

const slugRandomChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomSlug(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	out := make([]byte, n)
	for i, c := range buf {
		out[i] = slugRandomChars[int(c)%len(slugRandomChars)]
	}
	return string(out)
}

// --- katakana -> romaji ---
//
// Hepburn 式ベース。特殊記号:
//   - ッ: 次の子音を重ねる (キット → kitto)
//   - ー: 直前の母音を伸ばす (ケーキ → keeki)
//   - ン: n (母音を後に伴う場合の n' は簡略化のため常に n)
//
// 未対応の kana は空文字扱いにする (slug 用途では削除)。

var katakana2Romaji = map[string]string{
	"キャ": "kya", "キュ": "kyu", "キョ": "kyo",
	"シャ": "sha", "シュ": "shu", "ショ": "sho",
	"チャ": "cha", "チュ": "chu", "チョ": "cho",
	"ニャ": "nya", "ニュ": "nyu", "ニョ": "nyo",
	"ヒャ": "hya", "ヒュ": "hyu", "ヒョ": "hyo",
	"ミャ": "mya", "ミュ": "myu", "ミョ": "myo",
	"リャ": "rya", "リュ": "ryu", "リョ": "ryo",
	"ギャ": "gya", "ギュ": "gyu", "ギョ": "gyo",
	"ジャ": "ja", "ジュ": "ju", "ジョ": "jo",
	"ビャ": "bya", "ビュ": "byu", "ビョ": "byo",
	"ピャ": "pya", "ピュ": "pyu", "ピョ": "pyo",

	"ア": "a", "イ": "i", "ウ": "u", "エ": "e", "オ": "o",
	"カ": "ka", "キ": "ki", "ク": "ku", "ケ": "ke", "コ": "ko",
	"ガ": "ga", "ギ": "gi", "グ": "gu", "ゲ": "ge", "ゴ": "go",
	"サ": "sa", "シ": "shi", "ス": "su", "セ": "se", "ソ": "so",
	"ザ": "za", "ジ": "ji", "ズ": "zu", "ゼ": "ze", "ゾ": "zo",
	"タ": "ta", "チ": "chi", "ツ": "tsu", "テ": "te", "ト": "to",
	"ダ": "da", "ヂ": "ji", "ヅ": "zu", "デ": "de", "ド": "do",
	"ナ": "na", "ニ": "ni", "ヌ": "nu", "ネ": "ne", "ノ": "no",
	"ハ": "ha", "ヒ": "hi", "フ": "fu", "ヘ": "he", "ホ": "ho",
	"バ": "ba", "ビ": "bi", "ブ": "bu", "ベ": "be", "ボ": "bo",
	"パ": "pa", "ピ": "pi", "プ": "pu", "ペ": "pe", "ポ": "po",
	"マ": "ma", "ミ": "mi", "ム": "mu", "メ": "me", "モ": "mo",
	"ヤ": "ya", "ユ": "yu", "ヨ": "yo",
	"ラ": "ra", "リ": "ri", "ル": "ru", "レ": "re", "ロ": "ro",
	"ワ": "wa", "ヲ": "o", "ン": "n",
}

func katakanaToRomaji(katakana string) string {
	// hiragana は先に katakana へ寄せる (ぁ..ん を ァ..ン へ)
	kata := hiraganaToKatakana(katakana)

	runes := []rune(kata)
	var out strings.Builder
	i := 0
	for i < len(runes) {
		// 促音 ッ: 次の音の頭子音を重ねて自分は消える
		if runes[i] == 'ッ' && i+1 < len(runes) {
			nextRomaji := lookupRomaji(runes, i+1)
			if nextRomaji.romaji != "" {
				out.WriteByte(nextRomaji.romaji[0])
				out.WriteString(nextRomaji.romaji)
				i += 1 + nextRomaji.consumed
				continue
			}
		}
		// 長音 ー: 直前の母音を繰り返す
		if runes[i] == 'ー' {
			if s := out.String(); s != "" {
				last := s[len(s)-1]
				if isVowel(last) {
					out.WriteByte(last)
				}
			}
			i++
			continue
		}
		r := lookupRomaji(runes, i)
		if r.romaji != "" {
			out.WriteString(r.romaji)
			i += r.consumed
			continue
		}
		// 未対応の rune は skip
		i++
	}
	return out.String()
}

type romajiLookup struct {
	romaji   string
	consumed int
}

// lookupRomaji は runes[at] から 2 文字 (拗音) → 1 文字の順に katakana2Romaji を引く。
func lookupRomaji(runes []rune, at int) romajiLookup {
	if at+1 < len(runes) {
		two := string(runes[at : at+2])
		if v, ok := katakana2Romaji[two]; ok {
			return romajiLookup{v, 2}
		}
	}
	one := string(runes[at])
	if v, ok := katakana2Romaji[one]; ok {
		return romajiLookup{v, 1}
	}
	return romajiLookup{"", 1}
}

func isVowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// hiraganaToKatakana はひらがな (U+3041..U+3096) を対応するカタカナ (U+30A1..U+30F6) に寄せる。
// kagome の Reading はカタカナだが、直入力等でひらがなが来ても救えるようにする。
func hiraganaToKatakana(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x3041 && r <= 0x3096 {
			b.WriteRune(r + 0x60)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
