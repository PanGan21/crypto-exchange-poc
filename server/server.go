package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/PanGan21/crypto-exchange-poc/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"

	MarketETH Market = "ETH"

	// Just for development fake private key
	exchangePrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
)

type (
	OrderType string
	Market    string

	PlaceOrderRequest struct {
		Type   OrderType // limit or market
		UserId int64
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		Id        int64
		UserId    int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrder struct {
		UserId int64
		Price  float64
		Size   float64
		Id     int64
	}

	APIError struct {
		Error string
	}
)

func Start() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	pk5 := "395df67f0c2d2d9fe1ad08d1bc8b6627011959b79c53d7dd6a3536a33ab8a4fd"
	user5 := NewUser(pk5, 5)
	ex.Users[user5.Id] = user5

	pk7 := "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3"
	user7 := NewUser(pk7, 7)
	ex.Users[user7.Id] = user7

	pk6 := "e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52"
	user6 := NewUser(pk6, 6)
	ex.Users[user6.Id] = user6

	e.GET("/trades/:market", ex.handleGetTrades)
	e.GET("/order/:userId", ex.handleGetOrders)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bid", ex.handleGetBestBid)
	e.GET("/book/:market/ask", ex.handleGetBestAsk)

	e.POST("/order", ex.handlePlaceOrder)

	e.DELETE("/order/:id", ex.handleCancelOrder)

	buyerAddress := common.HexToAddress("0x28a8746e75304c0780E011BEd21C72cD78cd535E")
	buyerBalance, err := client.BalanceAt(context.Background(), buyerAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

	sellerAddress := common.HexToAddress("0x95cED938F7991cd0dFcb48F0a06a40FA1aF46EBC")
	sellerBalance, err := client.BalanceAt(context.Background(), sellerAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

	user6Address := common.HexToAddress("0x3E5e9111Ae8eB78Fe1CC3bb8915d5D461F3Ef9A9")
	user6Balance, err := client.BalanceAt(context.Background(), user6Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("buyerBalance", buyerBalance)
	fmt.Println("sellerBalance", sellerBalance)
	fmt.Println("user6Balance", user6Balance)

	e.Start(":3000")
}

type User struct {
	Id         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string, id int64) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		Id:         id,
		PrivateKey: pk,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Exchange struct {
	Client     *ethclient.Client
	mu         sync.RWMutex
	Users      map[int64]*User
	Orders     map[int64][]*orderbook.Order // user to his orders
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		Client:     client,
		Users:      make(map[int64]*User),
		Orders:     make(map[int64][]*orderbook.Order),
		PrivateKey: pk,
		orderbooks: orderbooks,
	}, nil
}

func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, APIError{Error: "orderbook not found"})
	}

	return c.JSON(http.StatusOK, ob.Trades)
}

type GetOrdersResponse struct {
	Asks []Order
	Bids []Order
}

func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return err
	}

	ex.mu.RLock()
	orderbookOrders := ex.Orders[int64(userId)]
	ordersResponse := &GetOrdersResponse{
		Asks: []Order{},
		Bids: []Order{},
	}
	for i := 0; i < len(orderbookOrders); i++ {
		if orderbookOrders[i].Limit == nil {
			fmt.Printf("the limit of the order is nil %+v\n", orderbookOrders[i])
			continue
		}
		order := Order{
			Id:        orderbookOrders[i].Id,
			UserId:    orderbookOrders[i].UserId,
			Price:     orderbookOrders[i].Limit.Price,
			Size:      orderbookOrders[i].Size,
			Timestamp: orderbookOrders[i].Timestamp,
			Bid:       orderbookOrders[i].Bid,
		}

		if order.Bid {
			ordersResponse.Bids = append(ordersResponse.Bids, order)
		} else {
			ordersResponse.Asks = append(ordersResponse.Asks, order)
		}

	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, ordersResponse)
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := &OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}
	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				Id:        order.Id,
				UserId:    order.UserId,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				Id:        order.Id,
				UserId:    order.UserId,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]

	matches := ob.PlaceMarketOrder(order)

	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	totalSizeFilled := 0.0
	sumPrice := 0.0
	for i := 0; i < len(matchedOrders); i++ {
		id := matches[i].Bid.Id
		limitUserId := matches[i].Bid.UserId

		if isBid {
			limitUserId = matches[i].Ask.UserId
			id = matches[i].Ask.Id
		}

		matchedOrders[i] = &MatchedOrder{
			Id:     id,
			UserId: limitUserId,
			Size:   matches[i].SizeFilled,
			Price:  matches[i].Price,
		}

		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price
	}

	avgPrice := sumPrice / float64(len(matches))

	log.Printf("filled market order => %d | size [%.2f] | avgPrice [%.2f]", order.Id, totalSizeFilled, avgPrice)

	newOrderMap := make(map[int64][]*orderbook.Order)
	ex.mu.Lock()
	for userId, orderbookOrders := range ex.Orders {
		for i := 0; i < len(orderbookOrders); i++ {
			// if the order is not filled place it in the map copy
			// this means that size of the order is 0
			if !orderbookOrders[i].IsFilled() {
				newOrderMap[userId] = append(newOrderMap[userId], orderbookOrders[i])
			}
		}
	}
	ex.Orders = newOrderMap
	ex.mu.Unlock()

	return matches, matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	// keep track of the user orders
	ex.mu.Lock()
	ex.Orders[order.UserId] = append(ex.Orders[order.UserId], order)
	ex.mu.Unlock()

	log.Printf("new LIMIT order => type [%t] | price [%.2f] | size [%.2f]", order.Bid, order.Limit.Price, order.Size)
	return nil
}

type PlaceOrderResponse struct {
	OrderId int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserId)

	// limit orders
	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
	}

	// market orders
	if placeOrderData.Type == MarketOrder {
		matches, matchedOrders := ex.handlePlaceMarketOrder(market, order)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}

		// Delete the orders of the user when filled
		for _, matchedOrder := range matchedOrders {
			fmt.Printf("Deleting => %+v\n", matchedOrder)
			userOrders := ex.Orders[matchedOrder.UserId]
			// if size is 0 order can be deleted
			for i := 0; i < len(userOrders); i++ {
				if userOrders[i].IsFilled() {
					if matchedOrder.Id == userOrders[i].Id {
						userOrders[i] = userOrders[len(userOrders)-1]
						userOrders = userOrders[:len(userOrders)-1]
					}

				}
			}
		}

		for _, userOrders := range ex.Orders {
			for i := 0; i < len(userOrders); i++ {
				userOrders[i] = userOrders[len(userOrders)-1]
				userOrders = userOrders[:len(userOrders)-1]
			}
		}

	}

	resp := &PlaceOrderResponse{
		OrderId: order.Id,
	}

	return c.JSON(200, resp)
}

type PriceResponse struct {
	Price float64
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Bids()) == 0 {
		return fmt.Errorf("the bids are empty")
	}

	bestBidPrice := ob.Bids()[0].Price

	pr := PriceResponse{
		Price: bestBidPrice,
	}

	return c.JSON(http.StatusOK, pr)

}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Asks()) == 0 {
		return fmt.Errorf("the asks are empty")
	}

	bestAskPrice := ob.Asks()[0].Price

	pr := PriceResponse{
		Price: bestAskPrice,
	}

	return c.JSON(http.StatusOK, pr)

}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)

	log.Println("order cancelled id =>", id)

	return c.JSON(200, map[string]any{"msg": "order deleted"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserId]
		if !ok {
			return fmt.Errorf("user not found %d", match.Ask.UserId)
		}

		toUser, ok := ex.Users[match.Bid.UserId]
		if !ok {
			return fmt.Errorf("user not found %d", match.Bid.UserId)
		}

		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		// exchangePublicKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePublicKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting key to ECDSA")
		// }
		// toAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

		amount := big.NewInt(int64(match.SizeFilled))

		transferETH(ex.Client, fromUser.PrivateKey, toAddress, amount)
	}

	return nil
}

func transferETH(client *ethclient.Client, fromPrivkey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()

	publicKey := fromPrivkey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivkey)
	if err != nil {
		log.Fatal(err)
	}

	return client.SendTransaction(ctx, signedTx)
}
