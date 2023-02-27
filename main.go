package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/PanGan21/crypto-exchange-poc/client"
	"github.com/PanGan21/crypto-exchange-poc/server"
)

const (
	maxOrders = 3
)

var (
	tick   = 2 * time.Second
	myAsks = make(map[float64]int64)
	myBids = make(map[float64]int64)
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		marketSellOrder := &client.PlaceOrderParams{
			UserId: 6,
			Bid:    false,
			Size:   1000,
		}

		_, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(err)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserId: 6,
			Bid:    true,
			Size:   1000,
		}

		_, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(err)
		}
		<-ticker.C

	}
}

func makeMarketSimple(c *client.Client) {
	ticker := time.NewTicker(tick)

	for {
		bestAsk, err := c.GetBestAsk()
		if err != nil {
			log.Println(err)
		}

		bestBid, err := c.GetBestBid()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("exchange spread", spread)

		if len(myBids) < maxOrders {
			bidLimit := &client.PlaceOrderParams{
				UserId: 7,
				Bid:    true,
				Price:  bestBid + 100,
				Size:   1000,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(err)
			}

			myBids[bidLimit.Price] = bidOrderResp.OrderId
		}

		if len(myAsks) < maxOrders {

			askLimit := &client.PlaceOrderParams{
				UserId: 7,
				Bid:    false,
				Price:  bestAsk - 100,
				Size:   1000,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(err)
			}

			myAsks[askLimit.Price] = askOrderResp.OrderId
		}

		fmt.Println("best ask price", bestAsk)
		fmt.Println("best bid price", bestBid)

		<-ticker.C
	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserId: 5,
		Bid:    false,
		Price:  10_000,
		Size:   1_000_000,
	}

	bid := &client.PlaceOrderParams{
		UserId: 5,
		Bid:    true,
		Price:  9_000,
		Size:   1_000_000,
	}

	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}

	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	go server.Start()

	time.Sleep(1 * time.Second)

	pocClient := client.NewClient()

	if err := seedMarket(pocClient); err != nil {
		panic(err)
	}

	go makeMarketSimple(pocClient)

	time.Sleep(1 * time.Second)

	marketOrderPlacer(pocClient)

	// for {
	// limitOrderParams := &client.PlaceOrderParams{
	// 	UserId: 5,
	// 	Bid:    false,
	// 	Price:  10_000,
	// 	Size:   5_000_000,
	// }

	// _, err := pocClient.PlaceLimitOrder(limitOrderParams)
	// if err != nil {
	// 	panic(err)
	// }

	// otherLimitOrderParams := &client.PlaceOrderParams{
	// 	UserId: 6,
	// 	Bid:    false,
	// 	Price:  9_000,
	// 	Size:   500_000,
	// }

	// _, err = pocClient.PlaceLimitOrder(otherLimitOrderParams)
	// if err != nil {
	// 	panic(err)
	// }

	// buyLimitOrderParams := &client.PlaceOrderParams{
	// 	UserId: 6,
	// 	Bid:    true,
	// 	Price:  11_000,
	// 	Size:   500_000,
	// }

	// _, err = pocClient.PlaceLimitOrder(buyLimitOrderParams)
	// if err != nil {
	// 	panic(err)
	// }

	// marketOrderParam := &client.PlaceOrderParams{
	// 	UserId: 7,
	// 	Bid:    true,
	// 	Size:   1_000_000,
	// }

	// _, err = pocClient.PlaceMarketOrder(marketOrderParam)
	// if err != nil {
	// 	panic(err)
	// }

	// bestBidPrice, err := pocClient.GetBestBid()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("best bid price", bestBidPrice)

	// bestAskPrice, err := pocClient.GetBestAsk()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("best ask price", bestAskPrice)

	// time.Sleep(1 * time.Second)
	// }

	select {}
}
