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
	tick = 2 * time.Second
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		trades, err := c.GetTrades("ETH")
		if err != nil {
			panic(err)
		}

		if len(trades) > 0 {
			fmt.Printf("exchange price => %2.f\n", trades[len(trades)-1].Price)
		}

		orderMarketSellOrder := &client.PlaceOrderParams{
			UserId: 5,
			Bid:    false,
			Size:   1000,
		}
		_, err = c.PlaceMarketOrder(orderMarketSellOrder)
		if err != nil {
			log.Println(err)
		}

		marketSellOrder := &client.PlaceOrderParams{
			UserId: 6,
			Bid:    false,
			Size:   100,
		}

		_, err = c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(err)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserId: 6,
			Bid:    true,
			Size:   100,
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
		orders, err := c.GetOrders(7)
		if err != nil {
			log.Println(err)
		}

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

		if len(orders.Bids) < maxOrders {
			bidLimit := &client.PlaceOrderParams{
				UserId: 7,
				Bid:    true,
				Price:  bestBid + 100,
				Size:   1000,
			}

			_, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(err)
			}
		}

		if len(orders.Asks) < maxOrders {
			askLimit := &client.PlaceOrderParams{
				UserId: 7,
				Bid:    false,
				Price:  bestAsk - 100,
				Size:   1000,
			}

			_, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(err)
			}

		}

		fmt.Println("best ask price", bestAsk)
		fmt.Println("best bid price", bestBid)

		<-ticker.C
	}
}

const userId = 5

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserId: userId,
		Bid:    false,
		Price:  10_000,
		Size:   1_000,
	}

	bid := &client.PlaceOrderParams{
		UserId: userId,
		Bid:    true,
		Price:  9_000,
		Size:   10_000,
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

	select {}
}
