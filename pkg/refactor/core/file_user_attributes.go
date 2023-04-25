package core

import (
	"sort"
	"strings"
)

const userAttributesKey = "user_attributes"

// AppendFileUserAttributes is a helper function to append user attributed to file annotation data. User attributes
// are simply comma separated string of attributes words.
func (data FileAnnotationsData) AppendFileUserAttributes(newAttributes []string) FileAnnotationsData {
	oldAttrsString := data.GetEntry(userAttributesKey)
	oldAttrsList := strings.Split(oldAttrsString, ",")

	var newAttrsList []string
	newAttrsList = append(newAttrsList, oldAttrsList...)
	newAttrsList = append(newAttrsList, newAttributes...)

	newAttrs := compileUserAttributesList(newAttrsList, nil)

	return data.SetEntry(userAttributesKey, newAttrs)
}

// ReduceFileUserAttributes is a helper function to remove (if exists) user attributed from file annotation data. User attributes
// are simply comma separated string of attributes words.
func (data FileAnnotationsData) ReduceFileUserAttributes(attributes []string) FileAnnotationsData {
	oldAttrsString := data.GetEntry(userAttributesKey)
	oldAttrsList := strings.Split(oldAttrsString, ",")
	newAttrs := compileUserAttributesList(oldAttrsList, attributes)

	return data.SetEntry(userAttributesKey, newAttrs)
}

// compileUserAttributesList creates a comma separated user attributes string from both a list and an excludes. The
// returned string should be clean of whitespaces and no repetitions.
func compileUserAttributesList(list []string, excludes []string) string {
	// create a map with all the values as key
	uniqMap := make(map[string]bool)
	for _, v := range list {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			uniqMap[v] = true
		}
	}

	// remove excludes from the map
	for _, v := range excludes {
		v = strings.TrimSpace(v)
		delete(uniqMap, v)
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	sort.Strings(uniqSlice)

	return strings.Join(uniqSlice, ",")
}
