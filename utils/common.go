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
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// Parase image prefix,image name,image tag. eg: percona/percona-server:5.7.35 -> percona,percona-server,5.7.35
func ParseImageName(image string) (string, string, string, error) {
	imagePrefix, imageString := filepath.Split(image)
	imagePrefix = strings.TrimSuffix(imagePrefix, "/")
	imageNameArry := strings.Split(imageString, ":")
	if len(imageNameArry) <= 1 {
		return "", "", "", fmt.Errorf("image name or tag is empty")
	}
	imageName := imageNameArry[0]
	imageTag := imageNameArry[1]
	return imagePrefix, imageName, imageTag, nil
}

// escapeJSONPointer encodes '~' and '/' according to RFC 6901.
var escapeJSONPointer = strings.NewReplacer(
	"~", "~0",
	"/", "~1",
).Replace

// JSON6902 represents a JSON Patch according to RFC 6902; the same as
// k8s.io/apimachinery/pkg/types.JSONPatchType.
type JSON6902 []interface{}

// NewJSONPatch creates a new JSON Patch according to RFC 6902; the same as
// k8s.io/apimachinery/pkg/types.JSONPatchType.
func NewJSONPatch() *JSON6902 { return &JSON6902{} }

func (*JSON6902) pointer(tokens ...string) string {
	var b strings.Builder

	for _, t := range tokens {
		_ = b.WriteByte('/')
		_, _ = b.WriteString(escapeJSONPointer(t))
	}

	return b.String()
}

// Add appends an "add" operation to patch.
//
// > The "add" operation performs one of the following functions,
// > depending upon what the target location references:
// >
// > o  If the target location specifies an array index, a new value is
// >    inserted into the array at the specified index.
// >
// > o  If the target location specifies an object member that does not
// >    already exist, a new member is added to the object.
// >
// > o  If the target location specifies an object member that does exist,
// >    that member's value is replaced.
func (patch *JSON6902) Add(path ...string) func(value interface{}) *JSON6902 {
	i := len(*patch)
	f := func(value interface{}) *JSON6902 {
		(*patch)[i] = map[string]interface{}{
			"op":    "add",
			"path":  patch.pointer(path...),
			"value": value,
		}
		return patch
	}

	*patch = append(*patch, f)

	return f
}

// Remove appends a "remove" operation to patch.
//
// > The "remove" operation removes the value at the target location.
// >
// > The target location MUST exist for the operation to be successful.
func (patch *JSON6902) Remove(path ...string) *JSON6902 {
	*patch = append(*patch, map[string]interface{}{
		"op":   "remove",
		"path": patch.pointer(path...),
	})

	return patch
}

// Replace appends a "replace" operation to patch.
//
// > The "replace" operation replaces the value at the target location
// > with a new value.
// >
// > The target location MUST exist for the operation to be successful.
func (patch *JSON6902) Replace(path ...string) func(value interface{}) *JSON6902 {
	i := len(*patch)
	f := func(value interface{}) *JSON6902 {
		(*patch)[i] = map[string]interface{}{
			"op":    "replace",
			"path":  patch.pointer(path...),
			"value": value,
		}
		return patch
	}

	*patch = append(*patch, f)

	return f
}

// Bytes returns the JSON representation of patch.
func (patch JSON6902) Bytes() ([]byte, error) { return patch.Data(nil) }

// Data returns the JSON representation of patch.
func (patch JSON6902) Data(client.Object) ([]byte, error) { return json.Marshal(patch) }

// IsEmpty returns true when patch has no operations.
func (patch JSON6902) IsEmpty() bool { return len(patch) == 0 }

// Type returns k8s.io/apimachinery/pkg/types.JSONPatchType.
func (patch JSON6902) Type() types.PatchType { return types.JSONPatchType }

// Merge7386 represents a JSON Merge Patch according to RFC 7386; the same as
// k8s.io/apimachinery/pkg/types.MergePatchType.
type Merge7386 map[string]interface{}

// NewMergePatch creates a new JSON Merge Patch according to RFC 7386; the same
// as k8s.io/apimachinery/pkg/types.MergePatchType.
func NewMergePatch() *Merge7386 { return &Merge7386{} }

// Add modifies patch to indicate that the member at path should be added or
// replaced with value.
//
// > If the provided merge patch contains members that do not appear
// > within the target, those members are added.  If the target does
// > contain the member, the value is replaced.  Null values in the merge
// > patch are given special meaning to indicate the removal of existing
// > values in the target.
func (patch *Merge7386) Add(path ...string) func(value interface{}) *Merge7386 {
	position := *patch

	for len(path) > 1 {
		p, ok := position[path[0]].(Merge7386)
		if !ok {
			p = Merge7386{}
			position[path[0]] = p
		}

		position = p
		path = path[1:]
	}

	if len(path) < 1 {
		return func(interface{}) *Merge7386 { return patch }
	}

	f := func(value interface{}) *Merge7386 {
		position[path[0]] = value
		return patch
	}

	position[path[0]] = f

	return f
}

// Remove modifies patch to indicate that the member at path should be removed
// if it exists.
func (patch *Merge7386) Remove(path ...string) *Merge7386 {
	return patch.Add(path...)(nil)
}

// Bytes returns the JSON representation of patch.
func (patch Merge7386) Bytes() ([]byte, error) { return patch.Data(nil) }

// Data returns the JSON representation of patch.
func (patch Merge7386) Data(client.Object) ([]byte, error) { return json.Marshal(patch) }

// IsEmpty returns true when patch has no modifications.
func (patch Merge7386) IsEmpty() bool { return len(patch) == 0 }

// Type returns k8s.io/apimachinery/pkg/types.MergePatchType.
func (patch Merge7386) Type() types.PatchType { return types.MergePatchType }
