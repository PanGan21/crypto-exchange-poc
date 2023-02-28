package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PanGan21/crypto-exchange-poc/orderbook"
	"github.com/PanGan21/crypto-exchange-poc/server"
)

const url = "http://localhost:3000"

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

type PlaceOrderParams struct {
	UserId int64
	Bid    bool
	// Price only needed for placing LIMIT orders
	Price float64
	Size  float64
}

func (c *Client) GetTrades(market string) ([]*orderbook.Trade, error) {
	endpoint := fmt.Sprintf("%s/trades/%s", url, market)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	trades := []*orderbook.Trade{}
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, err
	}

	return trades, nil
}

func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserId: p.UserId,
		Type:   server.LimitOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	endpoint := url + "/order"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

func (c *Client) GetOrders(userId int64) (*server.GetOrdersResponse, error) {
	endpoint := fmt.Sprintf("%s/order/%d", url, userId)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	orders := server.GetOrdersResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}
	return &orders, nil
}

func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserId: p.UserId,
		Type:   server.MarketOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	endpoint := url + "/order"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

func (c *Client) GetBestBid() (float64, error) {
	endpoint := fmt.Sprintf("%s/book/ETH/bid", url)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	response, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	priceResp := &server.PriceResponse{}
	if err := json.NewDecoder(response.Body).Decode(priceResp); err != nil {
		return 0, err
	}

	return priceResp.Price, nil
}

func (c *Client) GetBestAsk() (float64, error) {
	endpoint := fmt.Sprintf("%s/book/ETH/ask", url)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	response, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	priceResp := &server.PriceResponse{}
	if err := json.NewDecoder(response.Body).Decode(priceResp); err != nil {
		return 0, err
	}

	return priceResp.Price, nil
}

func (c *Client) CancelOrder(orderId int64) error {
	endpoint := fmt.Sprintf("%s/order/%d", url, orderId)

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	if err != nil {
		return err
	}

	return nil
}
