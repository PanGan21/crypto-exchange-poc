package main

import (
	"time"

	"github.com/PanGan21/crypto-exchange-poc/client"
	"github.com/PanGan21/crypto-exchange-poc/server"
)

func main() {
	go server.Start()

	time.Sleep(1 * time.Second)

	pocClient := client.NewClient()

	for {
		limitOrderParams := &client.PlaceOrderParams{
			UserId: 5,
			Bid:    false,
			Price:  10_000,
			Size:   500_000,
		}

		_, err := pocClient.PlaceLimitOrder(limitOrderParams)
		if err != nil {
			panic(err)
		}

		otherLimitOrderParams := &client.PlaceOrderParams{
			UserId: 6,
			Bid:    false,
			Price:  9_000,
			Size:   500_000,
		}

		_, err = pocClient.PlaceLimitOrder(otherLimitOrderParams)
		if err != nil {
			panic(err)
		}

		marketOrderParam := &client.PlaceOrderParams{
			UserId: 7,
			Bid:    true,
			Size:   1_000_000,
		}

		_, err = pocClient.PlaceMarketOrder(marketOrderParam)
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	}
}
