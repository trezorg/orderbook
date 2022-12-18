package orderbook

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

// BuyOrder represents the request to buy a crypto active
type BuyOrder struct {
	Amount uint
}

// SellOrder represents the request to sell a crypto active
type SellOrder struct {
	Number uint
	Price  uint
}

// Amount represents the total price to sell
func (so SellOrder) Amount() uint {
	return so.Number * so.Price
}

// Momento pattern
type RollbackCommand func()
type History []RollbackCommand

// Rollback undo commands in reverse order
func (h History) Rollback() {
	for i := len(h) - 1; i >= 0; i-- {
		h[i]()
	}
}

type OrdersHeap []SellOrder

func parentIdx(idx int) int {
	rest, div := idx%2, idx/2
	if rest == 0 {
		div--
	}
	if div < 0 {
		div = 0
	}
	return div
}

func (h OrdersHeap) heapify() {
	firstParent := (len(h) - 1) / 2
	for i := firstParent; i >= 0; i-- {
		h.down(i)
	}
}

func (h OrdersHeap) up(idx int) int {
	for idx >= 0 {
		parent := parentIdx(idx)
		if parent != idx && h.Less(idx, parent) {
			h[idx], h[parent] = h[parent], h[idx]
			idx = parent
		} else {
			break
		}
	}
	return idx
}

func (h OrdersHeap) down(idx int) {
	for idx < len(h) {
		childIdx := idx
		leftChildIdx := idx*2 + 1
		rightChildIdx := idx*2 + 2
		if leftChildIdx < len(h) && h.Less(leftChildIdx, childIdx) {
			childIdx = leftChildIdx
		}
		if rightChildIdx < len(h) && h.Less(rightChildIdx, childIdx) {
			childIdx = rightChildIdx
		}
		if childIdx != idx {
			h[idx], h[childIdx] = h[childIdx], h[idx]
			idx = childIdx
		} else {
			break
		}
	}
}

func (h *OrdersHeap) Push(item SellOrder) int {
	*h = append(*h, item)
	return h.up(len(*h) - 1)
}

func (h OrdersHeap) Len() int { return len(h) }

// Less function for Min heap
func (h OrdersHeap) Less(i, j int) bool {
	if h[i].Price < h[j].Price {
		return true
	}
	if h[i].Price > h[j].Price {
		return false
	}
	return h[i].Number > h[j].Number
}

func (h *OrdersHeap) Pop() SellOrder {
	if len(*h) == 0 {
		panic("empty heap")
	}
	item := (*h)[0]
	(*h)[0] = (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	h.down(0)
	return item
}

func (h OrdersHeap) Pick() SellOrder {
	if len(h) == 0 {
		panic("empty heap")
	}
	return (h)[0]
}

func (h *OrdersHeap) List() []SellOrder {
	res := make([]SellOrder, len(*h))
	i := 0
	for len(*h) > 0 {
		res[i] = h.Pop()
		i++
	}
	return res
}

type Result struct {
	Amount uint
	Price  uint
}
type OrderBook struct {
	Amount uint
	book   OrdersHeap
}

func NewOrderBook(so ...SellOrder) OrderBook {
	var book OrdersHeap
	book = append(book, so...)
	book.heapify()
	amount := uint(0)
	for _, order := range so {
		amount += order.Number
	}
	return OrderBook{book: book, Amount: amount}
}

func (ob *OrderBook) Sell(so ...SellOrder) {
	for _, s := range so {
		ob.book.Push(s)
		ob.Amount += s.Number
	}
}

func (ob OrderBook) Len() int {
	return ob.book.Len()
}

func (ob *OrderBook) Buy(bo BuyOrder) (History, Result) {
	var hist History
	result := Result{}
	for ob.book.Len() > 0 && result.Amount < bo.Amount {
		order := ob.book.Pop()
		available := min(order.Number, bo.Amount-result.Amount)
		result.Amount += available
		result.Price += available * order.Price
		if order.Number > available {
			order.Number -= available
			idx := ob.book.Push(order)
			hist = append(hist, func(amount uint, idx int) func() {
				return func() {
					ob.book[idx].Number += available
					ob.Amount += available
					ob.book.up(idx)
				}
			}(available, idx))
		} else {
			hist = append(hist, func(order SellOrder) func() {
				return func() {
					ob.Sell(order)
				}
			}(order))
		}
	}
	ob.Amount -= result.Amount
	return hist, result
}
