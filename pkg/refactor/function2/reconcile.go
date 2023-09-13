package function2

type reconcileObject interface {
	id() string
	hash() string
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
		search[item.id()] = true
	}
	for _, item := range target {
		_, ok := search[item.id()]
		if ok {
			continue
		}
		result = append(result, item.id())
	}

	return result
}

func reconcileCreate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	keys := map[string]bool{}
	for _, item := range target {
		keys[item.id()] = true
	}
	for _, item := range src {
		_, ok := keys[item.id()]
		if ok {
			continue
		}
		result = append(result, item.id())
	}

	return result
}

func reconcileUpdate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	search := map[string]reconcileObject{}
	for _, item := range src {
		search[item.id()] = item
	}

	for _, item := range target {
		searchItem, ok := search[item.id()]
		if !ok {
			continue
		}
		if searchItem.hash() == item.hash() {
			continue
		}
		result = append(result, item.id())
	}

	return result
}
