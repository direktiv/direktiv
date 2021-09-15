package flow

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

type pagination struct {
	after  string
	first  int32
	before string
	last   int32
	order  *grpc.PageOrder
	filter *grpc.PageFilter
}

func (p *pagination) Before() *ent.Cursor {

	if p.before == "" {
		return nil
	}

	x := decodeCursor(p.before)
	return x

}

func (p *pagination) First() *int {

	if p.first <= 0 {
		return nil
	}

	var x int
	x = int(p.first)
	return &x

}

func (p *pagination) After() *ent.Cursor {

	if p.after == "" {
		return nil
	}

	x := decodeCursor(p.after)
	return x

}

func (p *pagination) Last() *int {

	if p.last <= 0 {
		return nil
	}

	var x int
	x = int(p.last)
	return &x

}

func getPagination(args *grpc.Pagination) (*pagination, error) {

	p := new(pagination)
	p.after = args.GetAfter()
	p.first = args.GetFirst()
	p.before = args.GetBefore()
	p.last = args.GetLast()
	p.order = args.GetOrder()
	p.filter = args.GetFilter()

	return p, nil

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
	Value(int) map[string]interface{}
	ID(int) string
}

func encodeCustomCursor(id string) string {

	return base64.StdEncoding.EncodeToString([]byte(id))

}

func decodeCustomCursor(cursor string) (string, error) {

	x, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", errors.New("bad cursor")
	}

	return string(x), nil

}

func (cp *customPagination) Paginate(req *pagination) (*cpdOutput, error) {

	o := new(cpdOutput)

	filter := req.filter
	if filter == nil {
		filter = new(grpc.PageFilter)
	}

	err := cp.data.Filter(filter)
	if err != nil {
		return nil, err
	}

	order := req.order
	if order == nil {
		order = new(grpc.PageOrder)
	}

	err = cp.data.Order(order)
	if err != nil {
		return nil, err
	}

	o.TotalCount = cp.data.Total()

	if o.TotalCount == 0 {
		return o, nil
	}

	ids := make([]string, 0)
	list := make([]map[string]interface{}, 0)
	for i := 0; i < cp.data.Total(); i++ {
		ids = append(ids, cp.data.ID(i))
		list = append(list, cp.data.Value(i))
	}

	beforeIdx := cp.data.Total() - 1
	before := req.before
	if before != "" {
		x, err := decodeCustomCursor(before)
		if err == nil {
			for i := beforeIdx; i >= 0; i-- {
				id := cp.data.ID(i)
				if id == x {
					beforeIdx = i - 1
					break
				}
			}
		}
	}
	list = list[:beforeIdx+1]
	ids = ids[:beforeIdx+1]

	afterIdx := 0
	after := req.after
	if after != "" {
		x, err := decodeCustomCursor(after)
		if err == nil {
			for i := afterIdx; i < cp.data.Total(); i++ {
				id := cp.data.ID(i)
				if id == x {
					afterIdx = i + 1
					break
				}
			}
		}
	}
	list = list[afterIdx:]
	ids = ids[afterIdx:]

	firstIdx := len(list)
	if req.First() != nil && *req.First() < firstIdx {
		firstIdx = *req.First()
	}

	lastIdx := 0
	if req.Last() != nil && len(list)-*req.Last() > lastIdx {
		lastIdx = len(list) - *req.Last()
	}

	if firstIdx <= lastIdx {
		list = list[:0]
		ids = ids[:0]
	} else {
		list = list[lastIdx:firstIdx]
		ids = ids[lastIdx:firstIdx]
	}

	if len(list) > 0 {

		if lastIdx > 0 || afterIdx > 0 {
			o.PageInfo.HasPreviousPage = true
		}

		if firstIdx < cp.data.Total() || beforeIdx < cp.data.Total()-1 {
			o.PageInfo.HasNextPage = true
		}

		o.PageInfo.StartCursor = ids[0]
		o.PageInfo.EndCursor = ids[len(ids)-1]

		for idx := range ids {
			o.Edges = append(o.Edges, cpdEdge{
				Cursor: ids[idx],
				Node:   list[idx],
			})
		}

	}

	return o, nil

}

type cpdOutput struct {
	TotalCount int           `json:"totalCount"`
	PageInfo   grpc.PageInfo `json:"pageInfo"`
	Edges      []cpdEdge     `json:"edges"`
}

type cpdEdge struct {
	Cursor string                 `json:"cursor"`
	Node   map[string]interface{} `json:"node"`
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

	if filter.GetField() != "" && filter.GetField() != "NAME" {
		return fmt.Errorf("invalid filter field: %s", filter.GetField())
	}

	// TODO
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

	if order.GetField() != "" && order.GetField() != "NAME" {
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
