package utils

import (
	"regexp"
	"strconv"
	"strings"

	rndStr "github.com/dchest/uniuri"
	"github.com/gobeam/stringy"
	"golang.org/x/exp/constraints"
	"k8s.io/utils/strings/slices"
)

func MustGenerateAcronym(name string, maxSize int, filter []string) string {
	if acron := GenerateAcronym(name, maxSize, 9, filter); acron != "" {
		return acron
	}

	for startSize := Min(maxSize, 3); startSize <= maxSize; startSize++ {
		count, retry := 0, 100
		for random := rndStr.NewLen(startSize); count < retry; random = rndStr.NewLen(startSize) {
			count++
			if !slices.Contains(filter, random) {
				return strings.ToUpper(random)
			}
		}
	}
	return strings.ToUpper(rndStr.NewLen(maxSize))
}

func GenerateAcronym(name string, maxSize int, versionsPerOption int, filter []string) string {

	snakeName := stringy.New(name).SnakeCase().ToUpper()
	words := strings.Split(snakeName, "_")

	//clean words that are made just of numbers
	words = slices.Filter(nil, words, func(s string) bool {
		return !regexp.MustCompile(`^\d+$`).MatchString(s)
	})

	lastWord := words[len(words)-1]
	if acron := wordAcronym(lastWord, maxSize, versionsPerOption, filter); acron != "" {
		return acron
	}

	if len(words) > 1 {
		if acron := wordsAcronym(words, maxSize, versionsPerOption, filter); acron != "" {
			return acron
		}
	}
	return ""
}

func wordsAcronym(words []string, maxSize int, versionsPerOption int, filter []string) string {
	maxWords := Min(len(words), maxSize)
	var initials string
	startAt := len(words) - maxWords
	for i := startAt; i < len(words); i++ {
		initials += string(words[i][0])
		wordsLeft := len(words) - i - 1
		expectedSize := len(initials) + wordsLeft
		if expectedSize < maxSize {
			for j := 1; j <= maxSize-expectedSize; j++ {
				initials += string(words[i][j])
			}
		}
	}
	if acron := wordAcronym(initials, maxSize, versionsPerOption, filter); acron != "" {
		return acron
	}
	//try to start at higher index
	if startAt > 0 {
		return wordsAcronym(words[1:], maxSize, versionsPerOption, filter)
	}
	return ""
} 

func wordAcronym(wrd string, maxSize int, versionsPerOption int, filter []string) string {
	size := Min(len(wrd), maxSize)
	versionsPerOption = Min(versionsPerOption, 10)
	word := stringy.New(wrd)
	for i := 1; i < size; i++ {
		acron := word.Tease(i, "")
		if !slices.Contains(filter, acron) {
			return acron
		}
	}
	for i := 1; i < size-1; i++ {
		for j := 1; j < versionsPerOption; j++ {
			acron := word.Tease(i, strconv.Itoa(j))
			if !slices.Contains(filter, acron) {
				return acron
			}
		}
	}
	if maxSize > size {
		for i := 1; i < versionsPerOption; i++ {
			acron := word.Suffix(strconv.Itoa(i))
			if !slices.Contains(filter, acron) {
				return acron
			}
		}
	}
	return ""
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
