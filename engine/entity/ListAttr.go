package entity

import "github.com/xiaonanln/goworld/engine/gwlog"

// ListAttr is a attribute for a list of attributes
type ListAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	items  []interface{}
}

// Size returns size of ListAttr
func (a *ListAttr) Size() int {
	return len(a.items)
}

func (a *ListAttr) clearParent() {
	a.parent = nil
	a.pkey = nil

	a.clearOwner()
}

func (a *ListAttr) clearOwner() {
	a.owner = nil
	a.flag = 0
	a.path = nil

	// clear owner of children recursively
	for _, v := range a.items {
		if ma, ok := v.(*MapAttr); ok {
			ma.clearOwner()
		} else if la, ok := v.(*ListAttr); ok {
			la.clearOwner()
		}
	}
}

func (a *ListAttr) setParent(owner *Entity, parent interface{}, pkey interface{}, flag attrFlag) {
	a.parent = parent
	a.pkey = pkey

	a.setOwner(owner, flag)
}

func (a *ListAttr) setOwner(owner *Entity, flag attrFlag) {
	a.owner = owner
	a.flag = flag

	// set owner of children recursively
	for _, v := range a.items {
		if ma, ok := v.(*MapAttr); ok {
			ma.setOwner(owner, flag)
		} else if la, ok := v.(*ListAttr); ok {
			la.setOwner(owner, flag)
		}
	}
}

// Set sets item value
func (a *ListAttr) set(index int, val interface{}) {
	a.items[index] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrChangeToClients(index, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrChangeToClients(index, sa.ToList())
	} else {
		a.sendListAttrChangeToClients(index, val)
	}
}

func (a *ListAttr) sendListAttrChangeToClients(index int, val interface{}) {
	owner := a.owner
	if owner != nil {
		// send the change to owner's client
		owner.sendListAttrChangeToClients(a, index, val)
	}
}

func (a *ListAttr) sendListAttrPopToClients() {
	if owner := a.owner; owner != nil {
		owner.sendListAttrPopToClients(a)
	}
}

func (a *ListAttr) sendListAttrAppendToClients(val interface{}) {
	if owner := a.owner; owner != nil {
		owner.sendListAttrAppendToClients(a, val)
	}
}

func (a *ListAttr) getPathFromOwner() []interface{} {
	if a.path == nil {
		a.path = a._getPathFromOwner()
	}
	return a.path
}

func (a *ListAttr) _getPathFromOwner() []interface{} {
	path := make([]interface{}, 0, 4)
	if a.parent != nil {
		path = append(path, a.pkey)
		return getPathFromOwner(a.parent, path)
	}
	return path
}

// Get gets item value
func (a *ListAttr) get(index int) interface{} {
	return a.items[index]
}

// GetInt gets item value as int
func (a *ListAttr) GetInt(index int) int64 {
	return a.get(index).(int64)
}

// GetFloat gets item value as float64
func (a *ListAttr) GetFloat(index int) float64 {
	return a.get(index).(float64)
}

// GetStr gets item value as string
func (a *ListAttr) GetStr(index int) string {
	return a.get(index).(string)
}

// GetBool gets item value as bool
func (a *ListAttr) GetBool(index int) bool {
	return a.get(index).(bool)
}

// GetListAttr gets item value as ListAttr
func (a *ListAttr) GetListAttr(index int) *ListAttr {
	val := a.get(index)
	return val.(*ListAttr)
}

// GetMapAttr gets item value as MapAttr
func (a *ListAttr) GetMapAttr(index int) *MapAttr {
	val := a.get(index)
	return val.(*MapAttr)
}

// AppendInt puts int value to the end of list
func (a *ListAttr) AppendInt(v int64) {
	a.Append(v)
}

// AppendFloat puts float value to the end of list
func (a *ListAttr) AppendFloat(v float64) {
	a.Append(v)
}

// AppendBool puts bool value to the end of list
func (a *ListAttr) AppendBool(v bool) {
	a.Append(v)
}

// AppendStr puts string value to the end of list
func (a *ListAttr) AppendStr(v string) {
	a.Append(v)
}

// AppendMapAttr puts MapAttr value to the end of list
func (a *ListAttr) AppendMapAttr(attr *MapAttr) {
	a.Append(attr)
}

// AppendListAttr puts ListAttr value to the end of list
func (a *ListAttr) AppendListAttr(attr *ListAttr) {
	a.Append(attr)
}

// Pop removes the last item from the end
func (a *ListAttr) pop() interface{} {
	size := len(a.items)
	val := a.items[size-1]
	a.items = a.items[:size-1]

	if sa, ok := val.(*MapAttr); ok {
		sa.clearParent()
	} else if sa, ok := val.(*ListAttr); ok {
		sa.clearParent()
	}

	a.sendListAttrPopToClients()
	return val
}

func (a *ListAttr) PopInt() int64 {
	return a.pop().(int64)
}

func (a *ListAttr) PopFloat() float64 {
	return a.pop().(float64)
}

func (a *ListAttr) PopBool() bool {
	return a.pop().(bool)
}

func (a *ListAttr) PopStr() string {
	return a.pop().(string)
}

// PopListAttr removes the last item and returns as ListAttr
func (a *ListAttr) PopListAttr() *ListAttr {
	return a.pop().(*ListAttr)
}

// PopMapAttr removes the last item and returns as MapAttr
func (a *ListAttr) PopMapAttr() *MapAttr {
	return a.pop().(*MapAttr)
}

func (a *ListAttr) Del(val interface{}) {
	index := -1
	for i, val2 := range a.items {
		if val == val2 {
			index = i
			break
		}
	}

	if index >= 0 {
		a.items = append(a.items[:index], a.items[index+1:]...)
		// TODO sync client
	}
}

// append puts item to the end of list
func (a *ListAttr) Append(val interface{}) {
	a.items = append(a.items, val)
	index := len(a.items) - 1

	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in Append")
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrAppendToClients(sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in Append")
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrAppendToClients(sa.ToList())
	} else {
		a.sendListAttrAppendToClients(val)
	}
}

// SetInt sets int value at the index
func (a *ListAttr) SetInt(index int, v int64) {
	a.set(index, v)
}

// SetFloat sets float value at the index
func (a *ListAttr) SetFloat(index int, v float64) {
	a.set(index, v)
}

// SetBool sets bool value at the index
func (a *ListAttr) SetBool(index int, v bool) {
	a.set(index, v)
}

// SetStr sets string value at the index
func (a *ListAttr) SetStr(index int, v string) {
	a.set(index, v)
}

// SetMapAttr sets MapAttr value at the index
func (a *ListAttr) SetMapAttr(index int, attr *MapAttr) {
	a.set(index, attr)
}

// SetListAttr sets ListAttr value at the index
func (a *ListAttr) SetListAttr(index int, attr *ListAttr) {
	a.set(index, attr)
}

// ToList converts ListAttr to slice, recursively
func (a *ListAttr) ToList() []interface{} {
	l := make([]interface{}, len(a.items))

	for i, v := range a.items {
		if ma, ok := v.(*MapAttr); ok {
			l[i] = ma.ToMap()
		} else if la, ok := v.(*ListAttr); ok {
			l[i] = la.ToList()
		} else {
			l[i] = v
		}
	}
	return l
}

// AssignList assigns slice to ListAttr, recursively
func (a *ListAttr) AssignList(l []interface{}) {
	for _, v := range l {
		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.Append(ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			a.Append(ia)
		} else {
			a.Append(v)
		}
	}
}

func (a *ListAttr) ForEach(f func(index int, val interface{}) bool) {
	for i, v := range a.items {
		if !f(i, v) {
			break
		}
	}
}

// NewListAttr creates a new ListAttr
func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
