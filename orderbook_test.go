package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.RemoveOrder(buyOrderB)

	fmt.Println(l)
}
func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 5)
	ob.PlaceLimitsOrder(10_000, sellOrderA)
	ob.PlaceLimitsOrder(9_000, sellOrderB)

	assert(t, len(ob.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	// Provide liquidity
	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitsOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buydOrderA := NewOrder(true, 5)
	buydOrderB := NewOrder(true, 8)
	buydOrderC := NewOrder(true, 10)
	buydOrderD := NewOrder(true, 1)

	ob.PlaceLimitsOrder(5_000, buydOrderC)
	ob.PlaceLimitsOrder(5_000, buydOrderD)
	ob.PlaceLimitsOrder(9_000, buydOrderB)
	ob.PlaceLimitsOrder(10_000, buydOrderA)

	assert(t, ob.BidTotalVolume(), float64(10+8+5+1))

	sellOrder := NewOrder(false, 20)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), 4.0) // (10 + 8 + 5 + 1) - 20 = 4
	assert(t, len(matches), 3)          // 3 limits: 5_000, 9_000, 10_000
	assert(t, len(ob.bids), 1)

	fmt.Printf("%+v", matches)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()

	buyOrder := NewOrder(true, 4)

	ob.PlaceLimitsOrder(100000.0, buyOrder)
	assert(t, ob.BidTotalVolume(), 4.0)

	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)
}
