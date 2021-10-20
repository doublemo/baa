package segmenter

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/huichen/sego"
)

var (
	segmenter             sego.Segmenter
	dirtyWords            map[string]bool
	mutex                 sync.RWMutex
	defaultDictionaryPath string
	defaultDirtyPath      string
)

// Init 初始化分词器
func Init(dictionaryPath, dirtyPath string) error {
	defaultDictionaryPath = dictionaryPath
	defaultDirtyPath = dirtyPath
	dirtyWords = make(map[string]bool)

	segmenter.LoadDictionary(dictionaryPath)
	f, err := os.Open(dirtyPath)
	if err != nil {
		return err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		words := strings.Split(strings.ToUpper(strings.TrimSpace(scanner.Text())), " ") // 均处理为大写
		if words[0] != "" {
			dirtyWords[words[0]] = true
		}
	}

	return nil
}

// Reload 重新加载
func Reload(dictionaryPath, dirtyPath string) error {
	if dictionaryPath == "" {
		dictionaryPath = defaultDictionaryPath
	}

	if dirtyPath == "" {
		dirtyPath = defaultDirtyPath
	}

	segmenter.LoadDictionary(dictionaryPath)
	f, err := os.Open(dirtyPath)
	if err != nil {
		return err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	dirtyMap := make(map[string]bool)
	for scanner.Scan() {
		words := strings.Split(strings.ToUpper(strings.TrimSpace(scanner.Text())), " ") // 均处理为大写
		if words[0] != "" {
			dirtyMap[words[0]] = true
		}
	}

	mutex.Lock()
	dirtyWords = dirtyMap
	mutex.Unlock()
	return nil
}

// IsDirtyWords 检查是否有脏词
func IsDirtyWords(text string) bool {
	binText := []byte(text)
	segments := segmenter.Segment(binText)
	for _, seg := range segments {
		mutex.RLock()
		if dirtyWords[strings.ToUpper(string(binText[seg.Start():seg.End()]))] {
			mutex.RUnlock()
			return true
		}
		mutex.RUnlock()
	}
	return false
}

// ReplaceDirtyWords 替换脏话
func ReplaceDirtyWords(text, replaceWord string) (string, bool) {
	not := false
	bin := []byte(text)
	segments := segmenter.Segment(bin)
	cleanText := make([]byte, 0, len(bin))
	for _, seg := range segments {
		word := bin[seg.Start():seg.End()]
		mutex.RLock()
		if dirtyWords[strings.ToUpper(string(word))] {
			not = true
			cleanText = append(cleanText, []byte(strings.Repeat(replaceWord, utf8.RuneCount(word)))...)
		} else {
			cleanText = append(cleanText, word...)
		}
		mutex.RUnlock()
	}

	return string(cleanText), not
}

func AddDirtyWords(text string) {
	mutex.Lock()
	dirtyWords[text] = true
	mutex.Unlock()
}

func RemoveDirtyWords(text string) {
	mutex.Lock()
	delete(dirtyWords, text)
	mutex.Unlock()
}
