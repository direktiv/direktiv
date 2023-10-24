package service

// reconcileObject interface helps reconcile logic to identify objects and differences in lists.
type reconcileObject interface {
	GetID() string
	GetValueHash() string
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
		search[item.GetID()] = true
	}
	for _, item := range target {
		_, ok := search[item.GetID()]
		if ok {
			continue
		}
		result = append(result, item.GetID())
	}

	return result
}

func reconcileCreate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	keys := map[string]bool{}
	for _, item := range target {
		keys[item.GetID()] = true
	}
	for _, item := range src {
		_, ok := keys[item.GetID()]
		if ok {
			continue
		}
		result = append(result, item.GetID())
	}

	return result
}

func reconcileUpdate(src []reconcileObject, target []reconcileObject) []string {
	result := []string{}

	search := map[string]reconcileObject{}
	for _, item := range src {
		search[item.GetID()] = item
	}

	for _, item := range target {
		searchItem, ok := search[item.GetID()]
		if !ok {
			continue
		}
		if searchItem.GetValueHash() == item.GetValueHash() {
			continue
		}
		result = append(result, item.GetID())
	}

	return result
}
