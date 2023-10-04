package function

type reconcileObject interface {
	getID() string
	getValueHash() string
}

type reconcileResult struct {
	deletes []string
	creates []string
	updates []string
}

func reconcile(src []reconcileObject, target []reconcileObject) *reconcileResult {
	return &reconcileResult{
		deletes: reconcileDelete(src, target),
		creates: reconcileCreate(src, target),
		updates: reconcileUpdate(src, target),
	}
}

func reconcileDelete(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}
	search := map[string]bool{}
	for _, item := range src {
		search[item.getID()] = true
	}
	for _, item := range target {
		_, ok := search[item.getID()]
		if ok {
			continue
		}
		result = append(result, item.getID())
	}

	return result
}

func reconcileCreate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	keys := map[string]bool{}
	for _, item := range target {
		keys[item.getID()] = true
	}
	for _, item := range src {
		_, ok := keys[item.getID()]
		if ok {
			continue
		}
		result = append(result, item.getID())
	}

	return result
}

func reconcileUpdate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	search := map[string]reconcileObject{}
	for _, item := range src {
		search[item.getID()] = item
	}

	for _, item := range target {
		searchItem, ok := search[item.getID()]
		if !ok {
			continue
		}
		if searchItem.getValueHash() == item.getValueHash() {
			continue
		}
		result = append(result, item.getID())
	}

	return result
}
