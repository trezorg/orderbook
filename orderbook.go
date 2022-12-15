package orderbook

import (
	"container/heap"
)

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

func (h OrdersHeap) Len() int { return len(h) }

// Min heap
func (h OrdersHeap) Less(i, j int) bool { return h[i].Price < h[j].Price }
func (h OrdersHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *OrdersHeap) Push(x interface{}) {
	*h = append(*h, x.(SellOrder))
}

func (h *OrdersHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
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
	book := OrdersHeap(so)
	heap.Init(&book)
	amount := uint(0)
	for _, order := range so {
		amount += order.Number
	}
	return OrderBook{book: book, Amount: amount}
}

func (ob *OrderBook) Sell(so ...SellOrder) {
	for _, s := range so {
		heap.Push(&ob.book, s)
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
		order := heap.Pop(&ob.book).(SellOrder)
		available := min(order.Number, bo.Amount-result.Amount)
		order.Number -= available
		result.Amount += available
		result.Price += available * order.Price
		if order.Number > 0 {
			heap.Push(&ob.book, order)
			hist = append(hist, func() {
				ob.book[0].Number += available
				ob.Amount += available
			})
		} else {
			hist = append(hist, func() {
				order.Number += available
				ob.Sell(order)
			})
		}
	}
	ob.Amount -= result.Amount
	return hist, result
}
