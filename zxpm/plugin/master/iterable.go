package master

type Iterable[Idx comparable, Val any] interface {
	Next() bool
	Id() Idx
	Value() Val
	Len() int
}

type MapIterator[Idx comparable, Val any] struct {
	is   map[Idx]Val
	keys []Idx
	idx  int
}

func NewMapIterator[Idx comparable, Val any](
	is map[Idx]Val,
) *MapIterator[Idx, Val] {
	keys := make([]Idx, 0, len(is))
	for k := range is {
		keys = append(keys, k)
	}

	return &MapIterator[Idx, Val]{is, keys, -1}
}

func (i *MapIterator[Idx, Val]) Next() bool {
	if i.idx < len(i.keys) {
		i.idx++
	}
	return i.idx < len(i.keys)
}

func (i *MapIterator[Idx, Val]) Value() Val {
	return i.is[i.keys[i.idx]]
}

func (i *MapIterator[Idx, Val]) Id() Idx {
	return i.keys[i.idx]
}

func (i *MapIterator[Idx, Val]) Len() int {
	return len(i.keys)
}

type SliceIterator[Val any] struct {
	is  []Val
	idx int
}

func NewSliceIterator[Val any](
	is []Val,
) *SliceIterator[Val] {
	return &SliceIterator[Val]{is, -1}
}

func (i *SliceIterator[Val]) Next() bool {
	if i.idx < len(i.is) {
		i.idx++
	}
	return i.idx < len(i.is)
}

func (i *SliceIterator[Val]) Value() Val {
	return i.is[i.idx]
}

func (i *SliceIterator[Val]) Id() int {
	return i.idx
}

func (i *SliceIterator[Val]) Len() int {
	return len(i.is)
}
