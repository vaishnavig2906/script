package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/durianpay/dpay-common/api"
	"github.com/durianpay/dpay-common/dcerrors"
	"github.com/gojektech/heimdall/v6/httpclient"
)

type CheckVAStatusResponse struct {
	Id                       string    `json:"id"`
	PaymentID                string    `json:"payment_id"`
	CallbackVirtualAccountID string    `json:"callback_virtual_account_id"`
	ExternalID               string    `json:"external_id"`
	BankCode                 string    `json:"bank_code"`
	MerchantCode             string    `json:"merchant_code"`
	AccountNumber            string    `json:"account_number"`
	Amount                   int       `json:"amount"`
	TransactionTimestamp     time.Time `json:"transaction_timestamp"`
	SenderName               string    `json:"sender_name"`
}

const (
	URL = "https://api.xendit.co/callback_virtual_account_payments/payment_id=%s"
)

func CheckVAPaymentStatus(ctx context.Context, secretKey string, paymentID string) (resp CheckVAStatusResponse, dpayErr *dcerrors.DpayError) {
	url := fmt.Sprintf(URL, paymentID)
	timeout := 30000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	httpResponse, err := api.Get(ctx, url, getHTTPHeaders(secretKey), client)
	if err != nil {
		dpayErr = &dcerrors.DpayError{
			ErrorDescription: "error getting response from xendit",
			StatusCode:       http.StatusInternalServerError,
		}
		return
	}

	if httpResponse == nil {
		dpayErr = &dcerrors.DpayError{
			ErrorDescription: "no response from api, response is nil",
			StatusCode:       http.StatusInternalServerError,
		}
		fmt.Println(dpayErr.StatusCode)
		return
	}

	if httpResponse.StatusCode == http.StatusForbidden || httpResponse.StatusCode == http.StatusNotFound {
		dpayErr = &dcerrors.DpayError{
			ErrorDescription: "no response from api",
			StatusCode:       httpResponse.StatusCode,
		}
		fmt.Println(dpayErr.StatusCode)
		return
	}

	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		dpayErr = &dcerrors.DpayError{
			ErrorDescription: "error reading response body",
			StatusCode:       http.StatusInternalServerError,
		}
		return
	}

	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		dpayErr = &dcerrors.DpayError{
			ErrorDescription: "error decoding response",
			StatusCode:       http.StatusInternalServerError,
		}
		return
	}

	return
}

func getHTTPHeaders(secretKey string) (headers map[string]string) {
	basicAuth := fmt.Sprintf(
		"%s %s",
		"BASIC",
		base64.StdEncoding.EncodeToString([]byte(secretKey+":")),
	)

	headers = map[string]string{
		api.AuthorizationHeader: basicAuth,
		api.ContentTypeHeader:   api.ContentTypeJSON,
	}

	return
}

func main() {
	secretKey := os.Args[1]
	paymentID := os.Args[2]

	ctx := context.Background()

	resp, err := CheckVAPaymentStatus(ctx, secretKey, paymentID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp)
}
