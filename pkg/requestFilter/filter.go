package requestFilter

import (
	"encoding/json"
	"strconv"
)

func GetFilter(sort string, limit, page int, data []byte) (Filter, error) {
	var filterItems []FilterItem
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	if len(data) == 0 {
		return Filter{
			Limit:  uint(limit),
			Offset: uint(offset),
			Sort:   sort,
		}, nil
	}
	err := json.Unmarshal(data, &filterItems)
	if err != nil {
		return Filter{}, err
	}
	return Filter{
		Filters: filterItems,
		Limit:   uint(limit),
		Offset:  uint(offset),
		Sort:    sort,
	}, nil
}

func BuildFilter(params map[string]string, data []byte) (Filter, error) {
	var filterItems []FilterItem
	var filter Filter
	page := 0

	for param, value := range params {
		switch param {
		case "sort":
			filter.Sort = value
		case "group":
			filter.Group = value
		case "limit":
			limit, _ := strconv.Atoi(value)
			filter.Limit = uint(limit)
		case "page":
			page, _ = strconv.Atoi(value)
		case "trim":
			filter.Trim = true
		}
	}
	if page > 1 && filter.Limit > 0 {
		filter.Offset = uint(page-1) * (filter.Limit)
	}
	if len(data) == 0 {
		return filter, nil
	}
	err := json.Unmarshal(data, &filterItems)
	if err != nil {
		return Filter{}, err
	}
	filter.Filters = filterItems
	return filter, nil
}

func GetSimpleFilter(condition string, key string, value interface{}) Filter {
	return Filter{
		Filters: []FilterItem{
			GetSimpleFilterItem(condition, key, value),
		},
	}
}

func GetSimpleFilterItem(condition string, key string, value interface{}) FilterItem {
	return FilterItem{
		Condition: condition,
		Data: map[string]interface{}{
			key: value,
		},
	}
}

func (f *Filter) Append(condition string, key string, value interface{}) Filter {
	f.Filters = append(f.Filters, GetSimpleFilterItem(condition, key, value))
	return *f
}

func (f *Filter) ByInitiator(initiator int) Filter {
	f.Initiator = initiator
	return *f
}

func (f *Filter) ResetLimits() Filter {
	newFilter := Filter{
		Filters: f.Filters,
		Limit:   0,
		Offset:  0,
		Sort:    f.Sort,
		Group:   f.Group,
	}
	return newFilter
}

func (f *Filter) AppendGroup(operator string, items ...FilterItem) Filter {
	f.Filters = append(f.Filters, GetFilterGroup(operator, items...))
	return *f
}

func GetFilterGroup(operator string, items ...FilterItem) FilterItem {
	return FilterItem{
		Group: FilterItemGroup{
			Operator:    operator,
			FilterItems: items,
		},
	}
}

func (f *Filter) IsEmpty() bool {
	return len(f.Filters) == 0
}
