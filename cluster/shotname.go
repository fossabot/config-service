package cluster

import (
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"regexp"
	"strconv"
	"strings"

	rndStr "github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"github.com/gobeam/stringy"
	"k8s.io/utils/strings/slices"
)

//getAllShortNames returns the short names of all clusters for the customer in context
func getAllShortNames(c *gin.Context) []string {
	if clusters, err := mongo.GetAllForCustomerWithProjection(c, []types.Cluster{}, mongo.NewProjectionBuilder().
		ExcludeID().
		Include(utils.SHORT_NAME_FIELD).
		Get()); err != nil {
		utils.LogNTraceError("failed to read clusters", err, c)
		return nil
	} else {
		var shortNames []string
		for _, doc := range clusters {
			if doc.Attributes[utils.SHORT_NAME_ATTRIBUTE] != nil {
				shortNames = append(shortNames, doc.Attributes[utils.SHORT_NAME_ATTRIBUTE].(string))
			}
		}
		return shortNames
	}
}

//getUniqueShortName tries to create a short name from a long name and if it fails, it creates a random one
func getUniqueShortName(name string, c *gin.Context) string {
	maxSize := 5
	filter := getAllShortNames(c)
	if shortName := longName2short(name, maxSize, 9, filter); shortName != "" {
		return shortName
	}
	for startSize := utils.Min(maxSize, 3); startSize <= maxSize; startSize++ {
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

//longName2short tries to create a short name from a long name
func longName2short(name string, maxSize int, versionsPerOption int, filter []string) string {

	snakeName := stringy.New(name).SnakeCase().ToUpper()
	words := strings.Split(snakeName, "_")

	//clean words that are made just of numbers
	words = slices.Filter(nil, words, func(s string) bool {
		return !regexp.MustCompile(`^\d+$`).MatchString(s)
	})

	lastWord := words[len(words)-1]
	if shortName := shortName(lastWord, maxSize, versionsPerOption, filter); shortName != "" {
		return shortName
	}

	if len(words) > 1 {
		if shortName := slicedLongName2Short(words, maxSize, versionsPerOption, filter); shortName != "" {
			return shortName
		}
	}
	return ""
}

//slicedLongName2Short
func slicedLongName2Short(words []string, maxSize int, versionsPerOption int, filter []string) string {
	maxWords := utils.Min(len(words), maxSize)
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
	if shortName := shortName(initials, maxSize, versionsPerOption, filter); shortName != "" {
		return shortName
	}
	//try to start at higher index
	if startAt > 0 {
		return slicedLongName2Short(words[1:], maxSize, versionsPerOption, filter)
	}
	return ""
}

func shortName(wrd string, maxSize int, versionsPerOption int, filter []string) string {
	size := utils.Min(len(wrd), maxSize)
	versionsPerOption = utils.Min(versionsPerOption, 10)
	word := stringy.New(wrd)
	for i := 1; i < size; i++ {
		shortName := word.Tease(i, "")
		if !slices.Contains(filter, shortName) {
			return shortName
		}
	}
	for i := 1; i < size-1; i++ {
		for j := 1; j < versionsPerOption; j++ {
			shortName := word.Tease(i, strconv.Itoa(j))
			if !slices.Contains(filter, shortName) {
				return shortName
			}
		}
	}
	if maxSize > size {
		for i := 1; i < versionsPerOption; i++ {
			shortName := word.Suffix(strconv.Itoa(i))
			if !slices.Contains(filter, shortName) {
				return shortName
			}
		}
	}
	return ""
}