package orderbook

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

func TestOrderBookWithNotEnoughAmount(t *testing.T) {
	sellOrders := []SellOrder{
		{Price: 10, Number: 5},
		{Price: 20, Number: 4},
		{Price: 30, Number: 3},
		{Price: 40, Number: 2},
		{Price: 50, Number: 1},
	}
	totalPrice := uint(0)
	for _, order := range sellOrders {
		totalPrice += order.Price * order.Number
	}
	orderBook := NewOrderBook(sellOrders...)
	if orderBook.Len() != len(sellOrders) {
		t.Errorf("expected items in order book %v, got %v", len(sellOrders), orderBook.Len())
	}
	initialAvailable := orderBook.Amount
	buyOrder := BuyOrder{Amount: initialAvailable + 100}
	hist, result := orderBook.Buy(buyOrder)
	if result.Amount != initialAvailable {
		t.Errorf("expected bought items equal %v, got %v", initialAvailable, result.Amount)
	}
	if result.Price != totalPrice {
		t.Errorf("expected bought items price equal %v, got %v", totalPrice, result.Price)
	}
	if orderBook.Amount != 0 {
		t.Errorf("expected order book available amount 0, got %v", orderBook.Amount)
	}
	if orderBook.Len() != 0 {
		t.Errorf("expected order book length 0, got %v", orderBook.Len())
	}
	hist.Rollback()
	if orderBook.Amount != initialAvailable {
		t.Errorf("expected order book available amount after rollback %v, got %v", initialAvailable, orderBook.Amount)
	}
	if orderBook.Len() != len(sellOrders) {
		t.Errorf("expected order book length %v, got %v", len(sellOrders), orderBook.Len())
	}
}

func TestOrderBookWithEnoughAmount(t *testing.T) {
	sellOrders := []SellOrder{
		{Price: 10, Number: 5},
		{Price: 20, Number: 4},
		{Price: 30, Number: 3},
		{Price: 40, Number: 2},
		{Price: 50, Number: 1},
	}
	requiredAmount := uint(6)
	buyOrder := BuyOrder{Amount: requiredAmount}
	orderBook := NewOrderBook()
	orderBook.Sell(sellOrders...)
	initialAmount := orderBook.Amount
	hist, result := orderBook.Buy(buyOrder)
	if result.Amount != requiredAmount {
		t.Errorf("expected bought items equal %v, got %v", requiredAmount, result)
	}
	if orderBook.Amount != initialAmount-result.Amount {
		t.Errorf("expected order book available amount %v, got %v", orderBook.Amount, initialAmount-result.Amount)
	}
	if orderBook.Len() != 4 {
		t.Errorf("expected order book length 4, got %v", orderBook.Len())
	}
	expectedPrice := uint(5*10 + 20)
	if result.Price != expectedPrice {
		t.Errorf("expected bought items price equal %v, got %v", expectedPrice, result.Price)
	}
	hist.Rollback()
	if orderBook.Amount != initialAmount {
		t.Errorf("expected order book available amount after rollback %v, got %v", initialAmount, orderBook.Amount)
	}
	if orderBook.Len() != len(sellOrders) {
		t.Errorf("expected order book length %v, got %v", len(sellOrders), orderBook.Len())
	}
}

func TestOrderBookWithNoAmount(t *testing.T) {
	sellOrders := []SellOrder{
		{Price: 10, Number: 5},
		{Price: 20, Number: 4},
		{Price: 30, Number: 3},
		{Price: 40, Number: 2},
		{Price: 50, Number: 1},
	}
	buyOrder := BuyOrder{Amount: 0}
	orderBook := NewOrderBook(sellOrders...)
	hist, result := orderBook.Buy(buyOrder)
	if result.Amount != buyOrder.Amount {
		t.Errorf("expected bought items equal %v, got %v", buyOrder.Amount, result)
	}
	if len(hist) != 0 {
		t.Errorf("expected history length %v, got %v", 0, len(hist))
	}
	if orderBook.Len() != 5 {
		t.Errorf("expected order book length 4, got %v", orderBook.Len())
	}
}
func TestOrderBookWithNoOrders(t *testing.T) {
	buyOrder := BuyOrder{Amount: 1000}
	orderBook := NewOrderBook()
	hist, result := orderBook.Buy(buyOrder)
	if result.Amount != 0 {
		t.Errorf("expected bought items equal %v, got %v", 0, result)
	}
	if len(hist) != 0 {
		t.Errorf("expected history length %v, got %v", 0, len(hist))
	}
}
func BenchmarkOrderBook(b *testing.B) {

	randUint := func() uint {
		b := make([]byte, 8)
		rand.Read(b)
		return uint(binary.LittleEndian.Uint64(b))
	}

	var sellOrders []SellOrder
	for i := 0; i < 10000; i++ {
		sellOrders = append(sellOrders, SellOrder{
			Number: randUint(),
			Price:  randUint(),
		})
	}
	orderBook := NewOrderBook(sellOrders...)
	buyOrder := BuyOrder{Amount: randUint()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hist, _ := orderBook.Buy(buyOrder)
		hist.Rollback()
	}
}
