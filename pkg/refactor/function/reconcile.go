package function

type reconcileObject interface {
	getId() string
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
		search[item.getId()] = true
	}
	for _, item := range target {
		_, ok := search[item.getId()]
		if ok {
			continue
		}
		result = append(result, item.getId())
	}

	return result
}

func reconcileCreate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	keys := map[string]bool{}
	for _, item := range target {
		keys[item.getId()] = true
	}
	for _, item := range src {
		_, ok := keys[item.getId()]
		if ok {
			continue
		}
		result = append(result, item.getId())
	}

	return result
}

func reconcileUpdate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	search := map[string]reconcileObject{}
	for _, item := range src {
		search[item.getId()] = item
	}

	for _, item := range target {
		searchItem, ok := search[item.getId()]
		if !ok {
			continue
		}
		if searchItem.getValueHash() == item.getValueHash() {
			continue
		}
		result = append(result, item.getId())
	}

	return result
}
