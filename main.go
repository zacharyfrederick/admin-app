package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	uuid "github.com/satori/go.uuid"
)

type AdminApp struct {
	Wallet   *gateway.Wallet
	Gw       *gateway.Gateway
	Network  *gateway.Network
	Contract *gateway.Contract
}

type CreateFundRequest struct {
	Name          string `json: "name" binding: "required"`
	InceptionDate string `json: "inceptionDate" binding: "required"`
}

func main() {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("admin")

	if contract == nil {
		log.Fatalf("The contract could not be retrieved")
	}

	adminApp := &AdminApp{Wallet: wallet, Gw: gw, Contract: contract, Network: network}

	router := gin.Default()
	router.POST("/funds", adminApp.PostFundEndpoint)
	router.GET("/funds/:id", adminApp.GetFundByIdEndpoint)
	router.GET("/funds/:id/*action", GetInvestorsForFundEndpoint)

	router.Run()
}

func (a *AdminApp) PostFundEndpoint(c *gin.Context) {
	var createFundRequest CreateFundRequest

	err := c.BindJSON(&createFundRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameter"})
		return
	}

	validRequest := validateCreateFundRequest(&createFundRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted inceptionDate"})
	}

	fundId := uuid.NewV4().String()
	result, err := a.Contract.SubmitTransaction("CreateFund", fundId, createFundRequest.Name, createFundRequest.InceptionDate)
	if err != nil {
		log.Fatalf("Faiiled to submit transaction: %v", err)
		fmt.Println(result)
	}

	c.JSON(http.StatusOK, gin.H{"fundId": fundId})
}

func (a *AdminApp) GetFundByIdEndpoint(c *gin.Context) {
	fundId := c.Param("id")
	result, err := a.Contract.EvaluateTransaction("QueryFundById", fundId)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	c.JSON(http.StatusOK, string(result))
}

func GetInvestorsForFundEndpoint(c *gin.Context) {

}

func GetCapitalAccountsForFundEndpoint(c *gin.Context) {

}

func GetPortfoliosForFundEndpoint(c *gin.Context) {

}

func GetCapitalAccountActionsForFundEndpoint(c *gin.Context) {

}

func GetPortfolioActionsForFundEndpoint(c *gin.Context) {

}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func validateDate(date string) bool {
	return true
}

func validateCreateFundRequest(r *CreateFundRequest) bool {
	return true
}
