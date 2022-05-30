/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Min returns the smallest int64 that was passed in the arguments.
func Min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the largest int64 that was passed in the arguments.
func Max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// StringInArray check whether the str is in the strArray.
func StringInArray(str string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, str)
	return index < len(strArray) && strArray[index] == str
}

func GetOrdinal(name string) (int, error) {
	idx := strings.LastIndexAny(name, "-")
	if idx == -1 {
		return -1, fmt.Errorf("failed to extract ordinal from name: %s", name)
	}

	ordinal, err := strconv.Atoi(name[idx+1:])
	if err != nil {
		return -1, fmt.Errorf("failed to extract ordinal from name: %s", name)
	}
	return ordinal, nil
}

// Create the Update file.
func TouchUpdateFile() error {
	var err error
	var file *os.File

	if file, err = os.Create(FileIndicateUpdate); err != nil {
		return err
	}

	file.Close()
	return nil
}

// Remove the Update file.
func RemoveUpdateFile() error {
	return os.Remove(FileIndicateUpdate)
}

// Check update file exist.
func ExistUpdateFile() bool {
	f, err := os.Open(FileIndicateUpdate)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		return true
	}

	f.Close()
	return true
}

// Build the backup directory name by time.
func BuildBackupName(name string) (string, string) {
	cur_time := time.Now()
	return fmt.Sprintf("%s_%v%v%v%v%v%v", name, cur_time.Year(), int(cur_time.Month()),
			cur_time.Day(), cur_time.Hour(), cur_time.Minute(), cur_time.Second()),
		fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			cur_time.Year(), cur_time.Month(), cur_time.Day(),
			cur_time.Hour(), cur_time.Minute(), cur_time.Second())
}

func StringDiffIn(actual, desired []string) []string {
	diff := []string{}
	for _, aStr := range actual {
		// if is not in the desired list remove it.
		if _, exists := stringIn(aStr, desired); !exists {
			diff = append(diff, aStr)
		}
	}
	return diff
}

func stringIn(str string, strs []string) (int, bool) {
	for i, s := range strs {
		if s == str {
			return i, true
		}
	}
	return 0, false
}

func UnmarshalJSON(in io.Reader, obj interface{}) error {
	body, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("io read error: %s", err)
	}

	if err = json.Unmarshal(body, obj); err != nil {
		return fmt.Errorf("error unmarshal data, error: %s, body: %s", err, string(body))
	}
	return nil
}
