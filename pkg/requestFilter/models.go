package requestFilter

type FilterItem struct {
	Condition string
	Data      map[string]interface{}
	Group     FilterItemGroup
}

type FilterItemGroup struct {
	Operator    string
	FilterItems []FilterItem
}

type Filter struct {
	Filters   []FilterItem
	Limit     uint
	Offset    uint
	Sort      string
	Group     string
	Operator  string
	Trim      bool
	Initiator int
}
