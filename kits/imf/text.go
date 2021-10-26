package imf

import (
	"github.com/doublemo/baa/kits/imf/segmenter"
)

func filterText(text string, c FilterConfig) (string, bool) {
	return segmenter.ReplaceDirtyWords(text, c.TextReplaceWord)
}
