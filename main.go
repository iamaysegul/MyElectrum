package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ybbus/jsonrpc/v2"
)

var rpcUser = ""
var rpcPass = ""
var rpcClient jsonrpc.RPCClient

func init() {
	exec.Command("bash", "./init.sh").Output()
	
	cmd1, _ := exec.Command("electrum", "getconfig", "rpcuser", "--testnet").Output()
	rpcUser = strings.ReplaceAll(string(cmd1), "\n", "")

	cmd2, _ := exec.Command("electrum", "getconfig", "rpcpassword", "--testnet").Output()
	rpcPass = strings.ReplaceAll(string(cmd2), "\n", "")

	rpcClient = jsonrpc.NewClientWithOpts("http://127.0.0.1:7777/", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(rpcUser+":"+rpcPass)),
		},
	})

}

func main() {
	// fmt.Println(createWallet())
	// fmt.Println(loadWallet())
	// fmt.Println(getBalance())
	// fmt.Println(listAddresses())
	// fmt.Println(sendBtc("", "0.0001"))

	http.HandleFunc("/api/createWallet", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, createWallet())
	})

	http.HandleFunc("/api/loadWallet", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, loadWallet())
	})

	http.HandleFunc("/api/getBalance", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, getBalance())
	})

	http.HandleFunc("/api/listAddresses", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, listAddresses())
	})

	http.HandleFunc("/api/sendBtc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, sendBtc(r.URL.Query().Get("destination"), r.URL.Query().Get("amount")))
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Printf("Server Started: http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func createWallet() string {
	response, err := rpcClient.Call("create")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	responseString, _ := response.GetString()
	return responseString
}

func loadWallet() string {
	response, err := rpcClient.Call("load_wallet")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	return fmt.Sprintf("%v", response.Result)
}

func getBalance() string {
	type GetBalanceResponse struct {
		Confirmed   string `json:"confirmed"`
		Unconfirmed string `json:"unconfirmed"`
	}
	response, err := rpcClient.Call("getbalance")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	var getBalanceResponse *GetBalanceResponse
	response.GetObject(&getBalanceResponse)
	return fmt.Sprintf("confirmed: %v, unconfirmed: %v", getBalanceResponse.Confirmed, getBalanceResponse.Unconfirmed)
}

func listAddresses() string {
	response, err := rpcClient.Call("listaddresses")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	return fmt.Sprintf("%v", response.Result)
}

func sendBtc(destination string, amount string) string {
	type PayToParams struct {
		Destination string  `json:"destination"`
		Amount      float64 `json:"amount"`
	}
	var paytoParams PayToParams
	paytoParams.Destination = destination
	paytoParams.Amount, _ = strconv.ParseFloat(amount, 64)
	response, err := rpcClient.Call("payto", &paytoParams)
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	tx, err := response.GetString()
	type BroadcastParams struct {
		Tx string `json:"tx"`
	}
	var broadcastParams BroadcastParams
	broadcastParams.Tx = tx
	response, err = rpcClient.Call("broadcast", &broadcastParams)
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Message
	}
	responseString, _ := response.GetString()
	return responseString
}
