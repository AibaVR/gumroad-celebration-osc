package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AibaVR/vrc-osc-go/client"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Session struct {
	products    map[string]int
	client      *client.VRCOSC
	message     chan struct{}
	logger      *zap.Logger
	accessToken string
	param       string
}

func main() {
	// Init logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to init logger with err: %v", err)
	}

	// Load config
	err = godotenv.Load("env.local")
	if err != nil {
		logger.Fatal("Failed to load env vars. Please check your \"env.local\" file exists.")
	}

	c := client.NewVRCOSC()
	message := make(chan struct{})
	s := &Session{client: c, products: make(map[string]int), message: message, logger: logger, accessToken: os.Getenv("GUMROAD_ACCESS_TOKEN"), param: os.Getenv("VRC_PARAM")}

	logger.Info("Aiba's Gumroad celebration OSC has started!")

	go s.startPolling()
	go s.confettiListener()

	srvClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, os.Kill)
		<-sigint
		logger.Info("os kill signal received, exiting...")
		close(srvClosed)
	}()

	<-srvClosed
}

func (s *Session) sendConfetti() {
	logger := s.logger.With(zap.String("method", "sendConfetti"))

	logger.Info("Starting confetti")
	if err := s.client.SendAvatarParam(s.param, true); err != nil {
		logger.Error("failed sending param to VRChat", zap.Error(err))
	}

	time.Sleep(time.Second * 5)

	logger.Info("Stopping confetti")
	if err := s.client.SendAvatarParam(s.param, false); err != nil {
		logger.Error("failed sending param to VRChat", zap.Error(err))
	}
}

func (s *Session) confettiListener() {
	for {
		select {
		case <-s.message:
			s.sendConfetti()
			break
		}
	}
}

func (s *Session) pollGumroad() {
	logger := s.logger.With(zap.String("method", "pollGumroad"))
	logger.Info("Polling...")
	var jsonData = fmt.Sprintf(`{
		"access_token": "%s"
	}`, s.accessToken)
	request, err := http.NewRequest("GET", "https://api.gumroad.com/v2/products", bytes.NewBufferString(jsonData))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		logger.Error("error while polling gumroad", zap.Error(err))
	}
	c := &http.Client{}
	resp, err := c.Do(request)
	if err != nil {
		logger.Error("failed to send request", zap.Error(err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("bad response from gumroad", zap.String("status", resp.Status))
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("failed closing resp", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("error while decoding gumroad resp", zap.Error(err))
		return
	}

	var res ProductListResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logger.Error("error while decoding gumroad resp", zap.Error(err))
	}

	for _, product := range res.Products {
		count, ok := s.products[product.Id]
		if !ok {
			logger.Info("New product found, adding to list.", zap.Any("product", product.Name))
			s.products[product.Id] = product.SalesCount
			count = product.SalesCount
		}

		if count != product.SalesCount {
			diff := product.SalesCount - count
			for i := 0; i < diff; i++ {
				logger.Info("New sale!", zap.String("name", product.Name))
				s.message <- struct{}{}
			}
			s.products[product.Id] = product.SalesCount
		}
	}
}

func (s *Session) startPolling() {
	// Start the ticker to poll every minute
	ticker := time.NewTicker(time.Minute)
	for ; true; <-ticker.C {
		go s.pollGumroad()
	}
}

type ProductListResponse struct {
	Success  bool `json:"success"`
	Products []struct {
		Name                 string      `json:"name"`
		PreviewUrl           string      `json:"preview_url"`
		Description          string      `json:"description"`
		CustomizablePrice    bool        `json:"customizable_price"`
		RequireShipping      bool        `json:"require_shipping"`
		CustomReceipt        string      `json:"custom_receipt"`
		CustomPermalink      *string     `json:"custom_permalink"`
		SubscriptionDuration interface{} `json:"subscription_duration"`
		Id                   string      `json:"id"`
		Url                  interface{} `json:"url"`
		Price                int         `json:"price"`
		Currency             string      `json:"currency"`
		ShortUrl             string      `json:"short_url"`
		ThumbnailUrl         string      `json:"thumbnail_url"`
		Tags                 []string    `json:"tags"`
		FormattedPrice       string      `json:"formatted_price"`
		Published            bool        `json:"published"`
		ShownOnProfile       bool        `json:"shown_on_profile"`
		FileInfo             struct {
			Size string `json:"Size,omitempty"`
		} `json:"file_info"`
		MaxPurchaseCount interface{} `json:"max_purchase_count"`
		Deleted          bool        `json:"deleted"`
		CustomFields     []struct {
			Name     string `json:"name"`
			Required bool   `json:"required"`
			Type     string `json:"type"`
		} `json:"custom_fields"`
		CustomSummary      string      `json:"custom_summary"`
		IsTieredMembership bool        `json:"is_tiered_membership"`
		Recurrences        interface{} `json:"recurrences"`
		Variants           []struct {
			Title   string `json:"title"`
			Options []struct {
				Name             string      `json:"name"`
				PriceDifference  int         `json:"price_difference"`
				IsPayWhatYouWant bool        `json:"is_pay_what_you_want"`
				RecurrencePrices interface{} `json:"recurrence_prices"`
				Url              interface{} `json:"url"`
			} `json:"options"`
		} `json:"variants"`
		SalesCount    int     `json:"sales_count"`
		SalesUsdCents float64 `json:"sales_usd_cents"`
	} `json:"products"`
}
