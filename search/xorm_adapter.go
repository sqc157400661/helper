package search

import "github.com/go-xorm/xorm"

type XormCondition struct {
	XormPublic
	Join []*XormJoin
}

func (e *XormCondition) SetJoinOn(t, table, on string) Condition {
	if e.Join == nil {
		e.Join = make([]*XormJoin, 0)
	}
	join := &XormJoin{
		Type:       t,
		Table:      table,
		JoinOn:     on,
		XormPublic: XormPublic{},
	}
	e.Join = append(e.Join, join)
	return join
}

type XormJoin struct {
	Type   string
	Table  string
	JoinOn string
	XormPublic
}

func (e *XormJoin) SetJoinOn(t, table, on string) Condition {
	return nil
}

type XormPublic struct {
	Where map[string][]interface{}
	In    map[string][]interface{}
	Order []string
	Or    map[string][]interface{}
}

func (e *XormPublic) SetIn(k string, v []interface{}) {
	if e.In == nil {
		e.In = make(map[string][]interface{})
	}
	e.In[k] = v
}
func (e *XormPublic) SetWhere(k string, v []interface{}) {
	if e.Where == nil {
		e.Where = make(map[string][]interface{})
	}
	e.Where[k] = v
}

func (e *XormPublic) SetOr(k string, v []interface{}) {
	if e.Or == nil {
		e.Or = make(map[string][]interface{})
	}
	e.Or[k] = v
}

func (e *XormPublic) SetOrder(k string) {
	if e.Order == nil {
		e.Order = make([]string, 0)
	}
	e.Order = append(e.Order, k)
}

func MakeCondition(q interface{}, page ...Pagination) func(db *xorm.Session) *xorm.Session {
	return func(db *xorm.Session) *xorm.Session {
		condition := &XormCondition{
			XormPublic: XormPublic{},
		}
		if len(page) > 0 {
			db = db.Limit(page[0].GetPageSize(), page[0].GetPageSize()*(page[0].GetPageIndex()-1))
		}
		ResolveSearchQuery(q, condition, 0)
		for _, join := range condition.Join {
			if join == nil {
				continue
			}
			db = db.Join(join.Type, join.Table, join.JoinOn)
			for k, v := range join.Where {
				db = db.Where(k, v...)
			}
			for k, v := range join.In {
				db = db.In(k, v...)
			}
			for k, v := range join.Or {
				db = db.Or(k, v...)
			}
			for _, o := range join.Order {
				db = db.OrderBy(o)
			}
		}
		for k, v := range condition.Where {
			db = db.And(k, v...)
		}
		for k, v := range condition.In {
			db = db.In(k, v...)
		}
		for k, v := range condition.Or {
			db = db.Or(k, v...)
		}
		for _, o := range condition.Order {
			db = db.OrderBy(o)
		}
		return db
	}
}
