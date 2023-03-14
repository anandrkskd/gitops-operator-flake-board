package main

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ansiTime = `\(\d+\.\d+s\)`
const ansiPrefix = `---\s+FAIL:\s+kuttl/harness/`

// StripAnsi ...
func StripAnsi(str string, re *regexp.Regexp) string {
	return re.ReplaceAllString(str, "")
}

// Match ...
type Match struct {
	FileType  string   `json:"filename"`
	Context   []string `json:"context,omitempty"`
	MoreLines int      `json:"moreLines,omitempty"`
}

// TestFailEntry ...
type TestFailEntry struct {
	PRList   []int
	TestFail int
	LastSeen *time.Time
	LogURLs  map[int] /* pr number -> log urls */ []string
}

type TestFailEntryPriodic struct {
	PRList   []string
	TestFail int
	LastSeen *time.Time
	LogURLs  map[string] /* pr number -> log urls */ []string
}

// type periodicJobData struct {
// 	failure        string
// 	clusterVersion string
// 	url            string
// 	flag           bool
// }

// Result ...
type Result map[string]map[string][]Match

func main() {
	blobStorage, err := NewBlobStorage("./.cache")
	if err != nil {
		fmt.Println(err)
		return
	}

	testFailMap := map[string]TestFailEntry{}

	// store search results
	var result Result

	// jsonFile, err := os.Open("search.json")
	// if err != nil {
	// 	panic(err)
	// }
	// defer jsonFile.Close()
	runType := "pull"
	req, err := http.NewRequest("GET", "https://search.ci.openshift.org/search", nil)
	if err != nil {
		panic(err)
	}

	// https://search.ci.openshift.org/search?context=0&maxAge=336h&maxBytes=20971520&maxMatches=5&name=pull-ci-openshift-odo-main-&search=%5C%5BFail%5C%5D&type=build-log
	q := req.URL.Query()
	q.Add("search", "(?i)--- FAIL: kuttl/harness/1-")
	q.Add("maxAge", "336h")
	q.Add("context", "0")
	q.Add("type", "build-log")
	q.Add("name", "pull-ci-redhat-developer-gitops-operator-master-")
	q.Add("maxMatches", "5")
	q.Add("maxBytes", "20971520")
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		panic(err)
	}

	// fmt.Println("map:", string(byteValue))

	// iterate over all results
	for k, search := range result {
		expectedBuildLogURL, err := parseURL(k, runType)
		if err != nil {
			expectedBuildLogURL = ""
		}

		runTime, err := getTestJobRunTime(k, runType, *blobStorage)
		if err != nil {
			fmt.Printf("Error occurred on test log download: %v ", err)
			return
		}

		odoIndex := strings.Index(k, "redhat-developer_gitops-operator")

		prNumber := int64(-1)

		if odoIndex > -1 {
			str := k[odoIndex:]
			strArr := strings.Split(str, "/")

			prNumber, err = strconv.ParseInt(strArr[1], 10, 64)
		}

		// fmt.Printf("%s\n", file)
		// fmt.Println("map:", search)
		for _, matches := range search {
			// fmt.Printf("  %v\n", regexp)
			for _, match := range matches {
				lines := []string{}
				for _, line := range match.Context {
					// fmt.Printf("    %v\n", line)
					cleanLine := strings.TrimSpace(line)
					var re = regexp.MustCompile(ansiTime)
					cleanLine = StripAnsi(cleanLine, re)
					re = regexp.MustCompile(ansiPrefix)
					cleanLine = StripAnsi(cleanLine, re)
					// fmt.Println(cleanLine)
					// de-duplication
					// count each line only once
					dup := false
					for _, l := range lines {
						if l == cleanLine {
							dup = true
						}
					}
					if !dup {

						entry, exists := testFailMap[cleanLine]
						if !exists {
							entry = TestFailEntry{LogURLs: map[int][]string{}}
						}

						entry.TestFail++

						lines = append(lines, cleanLine)

						if runTime != nil {

							val := entry.LastSeen
							if val == nil {
								val = runTime
							}

							if runTime.After(*val) {
								val = runTime
							}
							entry.LastSeen = val

						}

						if prNumber >= 0 {

							matchFound := false
							for _, existingEntry := range entry.PRList {
								if int64(existingEntry) == prNumber {
									matchFound = true
								}
							}

							if !matchFound {
								entry.PRList = append(entry.PRList, int(prNumber))
							}

							// Add build log URL for the PR
							logURLList := entry.LogURLs[int(prNumber)]
							logURLList = append(logURLList, expectedBuildLogURL)
							entry.LogURLs[int(prNumber)] = logURLList
						}

						testFailMap[cleanLine] = entry

					}
					// fmt.Println(testFailMap)
				}
			}
		}

	}

	type TestFails struct {
		Score    int
		TestName string
		Fails    int
		LastSeen string
		PRList   []int
		Entry    TestFailEntry
	}

	// convert tests to slice so we can easily sort it
	fails := []TestFails{}
	for test, entry := range testFailMap {

		prList := entry.PRList

		sort.Sort(sort.Reverse(sort.IntSlice(prList)))

		lastSeenVal := ""

		// Score calculation
		score := 0
		{
			daysSinceLastSeen := 1

			lastSeenTime := entry.LastSeen
			if lastSeenTime != nil {

				//days := time.Now().Sub(*lastSeenTime).Hours() / 24
				days := time.Since(*lastSeenTime).Hours() / 24

				lastSeenVal = fmt.Sprintf("%d days ago", int(days))

				daysSinceLastSeen = int(days)
			}

			if daysSinceLastSeen == 0 {
				daysSinceLastSeen = 1
			}

			prListSize := len(prList)

			if prListSize > 6 {
				// >6 PRs does not imply any further strength than 6 PRs, for score calculation purposes.
				prListSize = 6
			}

			score = (10 * prListSize * entry.TestFail) / (daysSinceLastSeen)

			// fmt.Printf("%s %d %d\n", test, score, count)

			// Minimum score if there is at least one PR, and at least one fail, is 1
			if score == 0 && len(prList) > 0 && entry.TestFail > 0 {
				score = 1
			}
		}

		fails = append(fails, TestFails{TestName: test, Fails: entry.TestFail, PRList: prList, Score: score, LastSeen: lastSeenVal, Entry: entry})
	}

	sort.Slice(fails, func(i, j int) bool {
		one := fails[i].Score
		two := fails[j].Score

		// Primary sort: descending by score
		if one != two {
			return one > two
		}

		// Secondary sort: descending by fails
		one = fails[i].Fails
		two = fails[j].Fails
		if one != two {
			return one > two
		}

		// Tertiary sort: descring by pr list size
		one = len(fails[i].PRList)
		two = len(fails[j].PRList)
		if one != two {
			return one > two
		}

		// Finally, sort ascending by name
		return fails[j].TestName > fails[i].TestName
	})

	fmt.Println("# gitops-operator test statistics")
	// fmt.Printf("Last update: %s (UTC)\n\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	// fmt.Println("Generated with https://github.com/jgwest/odo-tools/ and https://github.com/kadel/odo-tools")
	fmt.Println("## FLAKY TESTS: Failed test scenarios in past 14 days")
	fmt.Println("| Failure Score<sup>*</sup> | Failures | Test Name | Last Seen | PR List and Logs ")
	fmt.Println("|---|---|---|---|---|")
	for _, f := range fails {

		// Skip failures that appear to be contained to a single PR
		if len(f.PRList) <= 1 {
			continue
		}

		prListString := fmt.Sprintf("%d: ", len(f.PRList))
		for _, prNumber := range f.PRList {

			logURLs := f.Entry.LogURLs[prNumber]

			prListString += fmt.Sprintf("[#%d](%s/%d)", prNumber, "https://github.com/redhat-developer/gitops-operator/pull/", prNumber)

			if len(logURLs) > 0 {

				prListString += "<sup>"

				for index, logURL := range logURLs {
					prListString += "[" + strconv.FormatInt(int64(index+1), 10) + "](" + logURL + ")"

					if index+1 != len(logURLs) {
						prListString += ", "
					}
				}

				prListString += "</sup>"
			}

			prListString += " "
		}

		fmt.Printf("| %d | %d | %s | %s | %s\n", f.Score, f.Fails, f.TestName, f.LastSeen, prListString)
	}

	fmt.Println()
	fmt.Println()
	// periodicjobstats()
	fmt.Println()
	fmt.Println("<sup>*</sup> - Failure score is an arbitrary severity estimate, and is approximately `(# of PRs the test failure was seen in * # of test failures) / (days since failure)`. See code for full algorithm -- PRs welcome for algorithm improvements.")
	fmt.Println()
	// fmt.Println("Graph represents the total no of testcase failures observed per day")
	// fmt.Println("![graph](https://gist.github.com/anandrkskd/1ea5606207f6141af21c7c3b0d527635/raw/graph.png)")
}

func getTestJobRunTime(url, runType string, blobStorage BlobStorage) (*time.Time, error) {
	urlContents, err := downloadTestLog(url, runType, blobStorage)
	if err != nil {
		return nil, err
	}

	contentsByLine := strings.Split(strings.Replace(urlContents, "\r\n", "\n", -1), "\n")

	if len(contentsByLine) == 0 {
		return nil, nil
	}

	// Parse the first line in the file to determine when the test started (and failed.)
	{
		topLine := contentsByLine[0]

		tok1 := strings.Split(topLine, " ")
		if len(tok1) == 0 {
			return nil, nil
		}

		// There's definitely a better way to parse this :P
		tok2 := strings.Split(tok1[0], "/")
		if len(tok2) < 3 {
			return nil, nil
		}

		year, err := strconv.ParseInt(tok2[0], 10, 32)
		if err != nil {
			return nil, err
		}

		month, err := strconv.ParseInt(tok2[1], 10, 32)
		if err != nil {
			return nil, err
		}

		day, err := strconv.ParseInt(tok2[2], 10, 32)
		if err != nil {
			return nil, err
		}

		result := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Now().Location())

		return &result, nil

	}

}

func parseURL(url, runType string) (string, error) {
	index := strings.LastIndex(url, "/")
	if index == -1 {
		return "", fmt.Errorf("parsing error")
	}

	index = strings.LastIndex(url[0:index-1], "/")
	if index == -1 {
		return "", fmt.Errorf("parsing error")
	}
	index = strings.LastIndex(url[0:index-1], "/")

	if runType == "pull" {
		return "https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator" + url[index:] + "/build-log.txt", nil
	} else if runType == "periodic" {
		return "https://storage.googleapis.com/origin-ci-test" + url[index:] + "/build-log.txt", nil
	}
	return "https://storage.googleapis.com/origin-ci-test" + url[index:] + "/build-log.txt", nil
}

func downloadTestLog(url, runType string, blobStorage BlobStorage) (string, error) {

	value, err := blobStorage.retrieve(url)
	if err != nil {
		return "", err
	}

	if value != "" {
		return value, nil
	}

	// convert
	// https://prow.svc.ci.openshift.org/view/gcs/origin-ci-test/pr-logs/pull/batch/pull-ci-openshift-odo-master-v4.2-integration-e2e-benchmark/2047
	// https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/redhat-developer_odo/5809/pull-ci-redhat-developer-odo-main-v4.10-integration-e2e/1541287908823011328
	// to
	// https://storage.googleapis.com/origin-ci-test/pr-logs/pull/batch/pull-ci-openshift-odo-master-v4.2-integration-e2e-benchmark/2047/build-log.txt
	// https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-openshift-odo-main-v4.8-operatorhub-integration-nightly/1429594453135331328/build-log.txt

	newURL, err := parseURL(url, runType)

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	err = blobStorage.store(url, string(byteValue))
	if err != nil {
		return "", err
	}

	return string(byteValue), nil

}

// BlobStorage ...
type BlobStorage struct {
	path string
}

// NewBlobStorage ...
func NewBlobStorage(pathParam string) (*BlobStorage, error) {
	blobStorage := BlobStorage{
		path: pathParam,
	}

	if _, err := os.Stat(pathParam); os.IsNotExist(err) {
		err = os.Mkdir(pathParam, 0755)
		if err != nil {
			return nil, err
		}
	}

	files, err := ioutil.ReadDir(pathParam)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		info, err := os.Stat(pathParam + "/" + f.Name())

		if err != nil {
			return nil, err
		}

		modTime := info.ModTime()

		diff := time.Now().Sub(modTime)

		// Delete cache entries older than 3 weeks
		if diff.Hours() > 24*7*3 {
			err := os.Remove(pathParam + "/" + f.Name())

			if err != nil {
				return nil, err
			}
		}

	}

	return &blobStorage, nil
}

func (s BlobStorage) store(key string, value string) error {
	base64Key := base32.StdEncoding.EncodeToString([]byte(key))

	expectedPath := s.path + "/" + base64Key[:18]

	err := ioutil.WriteFile(expectedPath, []byte(value), 0755)

	return err

}

func (s BlobStorage) retrieve(key string) (string, error) {
	base64Key := base32.StdEncoding.EncodeToString([]byte(key))

	expectedPath := s.path + "/" + base64Key[:18]

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		return "", nil
	}

	contents, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		return "", err
	}

	return string(contents), nil

}
