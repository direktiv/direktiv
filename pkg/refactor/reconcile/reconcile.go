package reconcile

// Item interface helps reconcile logic to identify objects and differences in lists.
type Item interface {
	GetID() string
	GetValueHash() string
}

type Result struct {
	Deletes []string
	Creates []string
	Updates []string
}

func Run(src []Item, target []Item) *Result {
	return &Result{
		Deletes: reconcileDelete(src, target),
		Creates: reconcileCreate(src, target),
		Updates: reconcileUpdate(src, target),
	}
}

func reconcileDelete(src []Item, target []Item) []string {
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

func reconcileCreate(src []Item, target []Item) []string {
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

func reconcileUpdate(src []Item, target []Item) []string {
	result := []string{}

	search := map[string]Item{}
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
