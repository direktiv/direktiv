package internallogger

import (
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

// filters the passed *LogMsgs if the given filter is supported if
// the given filter is not supported returns the input unfiltered.
func FilterLogmsg(filter *grpc.PageFilter, input []*LogMsgs) []*LogMsgs {
	res := input
	if filter.Field == "QUERY" && filter.Type == "MATCH" {
		res = filterMatchByWfStateIterator(filter.Val, input)
	}
	return res
}

// filters the input using the extracted values from the queryValue string.
// queryValue should be formatted like <workflow>::<state-id>::<loop-index>
// <state-id> and <indexId> is optional
// examples for queryValue:
// myworkflow or myworkflow:: or myworkflow::::
// myworkflow::getter or myworkflow::getter::
// myworkflow::getter::1
// ::getter::
// this method has two behaviors
// 1. if loop-index is left empty:
// when a logmsg from the input array has a matching pair of logtag values
// with the extracted values it will be added to the results
// 2: When the loop-index is provided:
// the result will contain all logmsg marked with the given
// loop-index starting the first match of the provided workflow and state-id
// additionally, all logmsgs from nested loops and childs will be added to the results.
func filterMatchByWfStateIterator(queryValue string, input []*LogMsgs) []*LogMsgs {
	values := strings.Split(queryValue, "::")
	state := ""
	workflow := ""
	iterator := ""
	if len(values) > 0 {
		workflow = values[0]
	}
	if len(values) > 1 {
		state = values[1]
	}
	if len(values) > 2 {
		iterator = values[2]
	}
	matchWf := make([]*LogMsgs, 0)
	matchState := make([]*LogMsgs, 0)
	matchIterator := make([]*LogMsgs, 0)
	for _, v := range input {
		if v.Tags["workflow"] == nil {
			v.Tags["workflow"] = ""
		}
		if v.Tags["state-id"] == nil {
			v.Tags["state-id"] = ""
		}
		if v.Tags["loop-index"] == nil {
			v.Tags["loop-index"] = ""
		}
		vWorkflow := fmt.Sprintf("%s", v.Tags["workflow"])
		vStateid := fmt.Sprintf("%s", v.Tags["state-id"])
		vLoopindex := fmt.Sprintf("%s", v.Tags["loop-index"])
		if v.Tags["workflow"] == workflow {
			matchWf = append(matchWf, v)
		}
		if vStateid == state &&
			workflow != "" && vWorkflow == workflow {
			matchState = append(matchState, v)
		}
		if vStateid == state &&
			workflow == "" {
			matchState = append(matchState, v)
		}
		if vStateid != "" && vStateid == state &&
			vWorkflow == workflow &&
			vLoopindex == iterator {
			matchIterator = append(matchIterator, v)
		}
		if vStateid == "" && vWorkflow == workflow &&
			vLoopindex == iterator {
			matchIterator = append(matchIterator, v)
		}
	}
	if state == "" && iterator == "" {
		return matchWf
	}
	if workflow == "" && iterator == "" {
		return matchState
	}
	if iterator != "" {
		if len(matchIterator) == 0 {
			return make([]*LogMsgs, 0)
		}
		if matchIterator[0].Tags["callpath"] == nil {
			matchIterator[0].Tags["callpath"] = ""
		}
		vcallpath := fmt.Sprintf("%v", matchIterator[0].Tags["callpath"])
		if matchIterator[0].Tags["instance-id"] == nil {
			matchIterator[0].Tags["instance-id"] = ""
		}
		vInstanceid := fmt.Sprintf("%v", matchIterator[0].Tags["instance-id"])
		callpath := AppendInstanceID(vcallpath, vInstanceid)
		childs := getAllChilds(callpath, input)
		originInstance := filterByInstanceId(fmt.Sprintf("%v", matchIterator[0].Tags["instance-id"]), input)
		subtree := append(originInstance, childs...)
		res := filterByIterrator(iterator, subtree)
		if nestedLoopHead := getNestedLoopHead(childs); nestedLoopHead != "" {
			nestedLoop := filterByInstanceId(nestedLoopHead, subtree)
			if len(nestedLoop) == 0 {
				return res
			}
			if nestedLoop[0].Tags["callpath"] == nil {
				nestedLoop[0].Tags["callpath"] = ""
			}
			if nestedLoop[0].Tags["instance-id"] == nil {
				nestedLoop[0].Tags["instance-id"] = ""
			}
			r := fmt.Sprintf("%s", nestedLoop[0].Tags["callpath"])
			c := fmt.Sprintf("%s", nestedLoop[0].Tags["instance-id"])
			callpath := AppendInstanceID(r, c)
			nestedLoopChilds := getAllChilds(callpath, subtree)
			nestedLoopSubtree := append(nestedLoop, nestedLoopChilds...)
			res = append(res, nestedLoopChilds...)
			res = append(res, nestedLoopSubtree...)
			res = removeDuplicate(res)
		}
		return res
	}
	return matchState
}

func filterByIterrator(iterator string, in []*LogMsgs) []*LogMsgs {
	res := make([]*LogMsgs, 0)
	if iterator == "" {
		return res
	}
	for _, v := range in {
		if v.Tags["loop-index"] == iterator {
			res = append(res, v)
		}
	}
	return res
}

func getNestedLoopHead(in []*LogMsgs) string {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return fmt.Sprintf("%v", v.Tags["instance-id"])
		}
	}
	return ""
}

func getAllChilds(callpath string, in []*LogMsgs) []*LogMsgs {
	res := make([]*LogMsgs, 0)
	for _, v := range in {
		if strings.HasPrefix(fmt.Sprintf("%v", v.Tags["callpath"]), callpath) {
			res = append(res, v)
		}
	}
	return res
}

func filterByInstanceId(instanceId string, in []*LogMsgs) []*LogMsgs {
	res := make([]*LogMsgs, 0)
	for _, v := range in {
		if v.Tags["instance-id"] == instanceId {
			res = append(res, v)
		}
	}
	return res
}

// https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicate(in []*LogMsgs) []*LogMsgs {
	allKeys := make(map[*LogMsgs]bool)
	list := []*LogMsgs{}
	for _, item := range in {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
