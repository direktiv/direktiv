package flow

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
)

type pagination struct {
	limit  int
	offset int
	order  []*grpc.PageOrder
	filter []*grpc.PageFilter
}

func getPagination(args *grpc.Pagination) (*pagination, error) {
	p := new(pagination)
	p.limit = int(args.GetLimit())
	p.offset = int(args.GetOffset())
	p.order = args.GetOrder()
	p.filter = args.GetFilter()

	return p, nil
}

func pageInfo(p *pagination, total int) *grpc.PageInfo {
	pi := new(grpc.PageInfo)
	pi.Limit = int32(p.limit)
	pi.Offset = int32(p.offset)
	pi.Total = int32(total)
	pi.Order = p.order
	pi.Filter = p.filter
	return pi
}

type customPagination struct {
	data customPaginationData
}

func newCustomPagination(cpd customPaginationData) *customPagination {
	cp := new(customPagination)

	cp.data = cpd

	return cp
}

const (
	paginationOrderingASC  = "ASC"
	paginationOrderingDESC = "DESC"
)

type customPaginationData interface {
	Filter(filter *grpc.PageFilter) error
	Order(order *grpc.PageOrder) error
	Total() int
	Value(idx int) map[string]interface{}
	ID(idx int) string
}

func (cp *customPagination) Paginate(req *pagination) (*cpdOutput, error) {
	o := new(cpdOutput)

	for _, f := range req.filter {
		if f == nil {
			f = new(grpc.PageFilter)
		}

		err := cp.data.Filter(f)
		if err != nil {
			return nil, err
		}
	}

	if len(req.order) > 1 {
		return nil, errors.New("cannot perform multiple ordering")
	}

	var order *grpc.PageOrder

	if len(req.order) == 0 {
		order = new(grpc.PageOrder)
	} else {
		order = req.order[0]
	}

	err := cp.data.Order(order)
	if err != nil {
		return nil, err
	}

	o.PageInfo.Total = int32(cp.data.Total())
	o.PageInfo.Filter = req.filter
	o.PageInfo.Limit = int32(req.limit)
	o.PageInfo.Offset = int32(req.offset)
	o.PageInfo.Order = req.order

	if o.PageInfo.Total == 0 {
		return o, nil
	}

	list := make([]map[string]interface{}, 0)
	for i := 0; i < cp.data.Total(); i++ {
		list = append(list, cp.data.Value(i))
	}

	if req.offset > 0 {
		if req.offset > len(list) {
			list = list[len(list):]
		} else {
			list = list[req.offset:]
		}
	}

	if req.limit > 0 {
		if req.limit <= len(list) {
			list = list[:req.limit]
		}
	}

	o.Results = list

	return o, nil
}

type cpdOutput struct {
	PageInfo grpc.PageInfo            `json:"pageInfo"`
	Results  []map[string]interface{} `json:"results"`
}

type cpdSecrets struct {
	list []string
}

func newCustomPaginationDataSecrets() *cpdSecrets {
	cpd := new(cpdSecrets)

	cpd.list = make([]string, 0)

	return cpd
}

func (cpds *cpdSecrets) Total() int {
	return len(cpds.list)
}

func (cpds *cpdSecrets) ID(idx int) string {
	return cpds.list[idx]
}

func (cpds *cpdSecrets) Value(idx int) map[string]interface{} {
	return map[string]interface{}{
		"name": cpds.list[idx],
	}
}

func (cpds *cpdSecrets) Filter(filter *grpc.PageFilter) error {
	if filter == nil {
		return nil
	}

	if filter.GetField() != "" && filter.GetField() != util.PaginationKeyName {
		return fmt.Errorf("invalid filter field: %s", filter.GetField())
	}

	switch filter.GetType() {
	case "":
	default:
		return fmt.Errorf("invalid filter type: %s", filter.GetType())
	}

	arg := filter.GetVal()

	secrets := make([]string, 0)

	for _, secret := range cpds.list {
		if strings.Contains(secret, arg) {
			secrets = append(secrets, secret)
		}
	}

	cpds.list = secrets

	return nil
}

func (cpds *cpdSecrets) Order(order *grpc.PageOrder) error {
	if order.GetField() != "" && order.GetField() != util.PaginationKeyName {
		return fmt.Errorf("invalid order field: %s", order.GetField())
	}

	sort.Strings(cpds.list)

	switch order.GetDirection() {
	case "":
		fallthrough
	case paginationOrderingASC:
	case paginationOrderingDESC:
		sort.Sort(sort.Reverse(sort.StringSlice(cpds.list)))
	default:
		return fmt.Errorf("invalid order direction: %s", order.GetDirection())
	}

	return nil
}

func (cpds *cpdSecrets) Add(name string) {
	cpds.list = append(cpds.list, name)
}

//type orderingInfo struct {
//	db           string
//	req          string
//	defaultOrder func(fields ...string) ent.OrderFunc
//	isDefault    bool
//}

//func (p *pagination) orderings(orderings []*orderingInfo) []ent.OrderFunc {
//	var fns []ent.OrderFunc
//
//	for _, o := range p.order {
//		var ordering *orderingInfo
//
//		for _, x := range orderings {
//			if x.req == o.Field {
//				ordering = x
//				break
//			}
//		}
//
//		if ordering == nil {
//			continue
//		}
//
//		direction := ordering.defaultOrder
//
//		if o.Direction == "ASC" {
//			direction = ent.Asc
//		} else if o.Direction == "DESC" {
//			direction = ent.Desc
//		}
//
//		field := ordering.db
//
//		fns = append(fns, direction(field))
//	}
//
//	if len(fns) == 0 {
//		for _, x := range orderings {
//			if x.isDefault {
//				fns = append(fns, x.defaultOrder(x.db))
//			}
//		}
//
//		if len(fns) == 0 {
//			fns = append(fns, orderings[0].defaultOrder(orderings[0].db))
//		}
//	}
//
//	return fns
//}

type filteringInfo struct {
	field string
	ftype string
}

//type entQuery[T, X any] interface {
//	Count(ctx context.Context) (int, error)
//	Order(o ...ent.OrderFunc) T
//	Limit(limit int) T
//	Offset(offset int) T
//	All(ctx context.Context) ([]X, error)
//}

//func entFilters[T, X any, Q entQuery[T, X]](p *pagination, filtersInfo map[*filteringInfo]func(query Q, v string) (Q, error)) []func(query Q) (Q, error) {
//	var filters []func(query Q) (Q, error)
//
//	for idx := range p.filter {
//		f := p.filter[idx]
//		var fn func(query Q, v string) (Q, error)
//
//		for k, x := range filtersInfo {
//			if k.field == f.Field && k.ftype == f.Type {
//				fn = x
//				break
//			}
//		}
//
//		if fn == nil {
//			continue
//		}
//
//		filters = append(filters, func(query Q) (Q, error) {
//			return fn(query, f.Val)
//		})
//	}
//
//	return filters
//}

//nolint:dupword
//func paginate[T, X any, Q entQuery[T, X]](
//	ctx context.Context,
//	params *grpc.Pagination,
//	q Q,
//	o []*orderingInfo,
//	f map[*filteringInfo]func(Q, string) (Q, error),
//) ([]X, *grpc.PageInfo, error) {
//	var err error
//
//	p, err := getPagination(params)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	orderings := p.orderings(o)
//	filters := entFilters[T, X](p, f)
//
//	for _, filter := range filters {
//		q, err = filter(q)
//		if err != nil {
//			return nil, nil, err
//		}
//	}
//
//	total, err := q.Count(ctx)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	if len(orderings) > 0 {
//		q = any(q.Order(orderings...)).(Q)
//	}
//
//	if p.limit > 0 {
//		q = any(q.Limit(p.limit)).(Q)
//	}
//
//	if p.offset > 0 {
//		q = any(q.Offset(p.offset)).(Q)
//	}
//
//	results, err := q.All(ctx)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return results, pageInfo(p, total), nil
//}
