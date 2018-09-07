package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	sdkClient "github.com/suddutt1/fabricgosdkclientcore"
)

func main() {
	configFile := ""
	flag.StringVar(&configFile, "config", "", "Please provide the client config yaml file")
	flag.Parse()
	if len(configFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	fmt.Println("Using config file ", configFile)
	sdkClient := new(sdkClient.FabricSDKClient)
	if !sdkClient.Init(configFile) {
		fmt.Printf("Unable to initialize the SDK Client with config file %s \n", configFile)
		os.Exit(1)
	}
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		renderOutput(200, "Service Available", c)
	})
	router.POST("/api/chaincode/:action", func(c *gin.Context) {
		jsonBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			renderOutput(200, "Unable to read the post body", c)
			return
		}
		jsonRequestMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBody, &jsonRequestMap)
		if err != nil {
			renderOutput(200, "Unable to json body", c)
			return
		}
		channel := getString(jsonRequestMap["channel"])
		ccid := getString(jsonRequestMap["ccid"])
		ccFn := getString(jsonRequestMap["fn"])
		user := getString(jsonRequestMap["user"])
		peers := getStringSlice(jsonRequestMap["peers"])
		args := buildArgsList(jsonRequestMap["args"])
		action := c.Param("action")
		if action == "invoke" {
			trxnOutput, isSuccess, err := sdkClient.InvokeTrxn(channel, user, ccid, ccFn, args, peers, nil)
			fmt.Println("Returned after invoke ", isSuccess)
			if !isSuccess || err != nil {
				renderOutput(200, fmt.Sprintf("Invoke failed with %+v", err), c)
				return
			}
			renderOutput(200, trxnOutput, c)
			return
		} else if action == "query" {
			trxnOutput, isSuccess, err := sdkClient.Query(channel, user, ccid, ccFn, args, peers, nil)
			fmt.Println("Returned after query ", isSuccess)
			if !isSuccess || err != nil {
				renderOutput(200, fmt.Sprintf("Query failed with %+v", err), c)
				return
			}
			renderOutput(200, trxnOutput, c)
			return
		}
		renderOutput(200, "Invalid action provided. Valid values are invokce|query. Access url pattern is api/chaincode/<action>", c)
	})

	router.POST("/api/admin/enrolladmin/:adminID", func(c *gin.Context) {
		adminID := c.Param("adminID")
		if !sdkClient.EnrollOrgAdmin(true, adminID) {
			renderOutput(200, fmt.Sprintf("Unable to enroll adminID %s ", adminID), c)
			return
		}
		renderOutput(200, fmt.Sprintf("AdminID %s enrolled successfully", adminID), c)
	})
	router.POST("/api/admin/enrolluser", func(c *gin.Context) {
		jsonBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			renderOutput(200, "Unable to read the post body", c)
			return
		}
		jsonRequestMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBody, &jsonRequestMap)
		if err != nil {
			renderOutput(200, "Unable to json body", c)
			return
		}
		userID := getString(jsonRequestMap["userId"])
		secret := getString(jsonRequestMap["secret"])
		org := getString(jsonRequestMap["org"])

		if !sdkClient.EnrollOrgUser(userID, secret, org) {
			renderOutput(200, fmt.Sprintf("Unable to enroll userID %s ", userID), c)
			return
		}
		renderOutput(200, fmt.Sprintf("UserID %s enrolled successfully", userID), c)
	})
	osSigChan := make(chan os.Signal)
	signal.Notify(osSigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-osSigChan
		fmt.Println("Ctrl-C detected..")
		sdkClient.Shutdown()
		os.Exit(0)

	}()
	router.Run(":8080")

}
func renderOutput(code int, payload interface{}, c *gin.Context) {
	resp := make(map[string]interface{})
	resp["isSuccess"] = true
	resp["ts"] = time.Now()
	if byteSlice, isOk := payload.([]byte); isOk {
		if obj, ok := isJSON(byteSlice); ok {
			resp["payload"] = obj
		} else {
			resp["payload"] = string(byteSlice)
		}

	} else {
		resp["payload"] = payload
	}
	c.JSON(code, resp)

}
func isJSON(bytes []byte) (interface{}, bool) {
	var genericIntfc interface{}
	if err := json.Unmarshal(bytes, &genericIntfc); err == nil {
		return genericIntfc, true
	}
	return nil, false
}
func getString(strIntfc interface{}) string {
	if strIntfc != nil {
		if str, ok := strIntfc.(string); ok {
			return str
		}
	}
	return ""
}
func getStringSlice(strIntfc interface{}) []string {
	if strIntfc != nil {

		if intfcSlice, ok := strIntfc.([]interface{}); ok {
			retList := make([]string, 0)
			for _, intfc := range intfcSlice {
				retList = append(retList, getString(intfc))
			}
			fmt.Printf("\nBefore return %+v", retList)
			return retList
		}
	}
	return make([]string, 0)
}
func buildArgsList(strIntfc interface{}) [][]byte {
	outputBytesSlice := make([][]byte, 0)
	strSlice := getStringSlice(strIntfc)
	for _, str := range strSlice {
		outputBytesSlice = append(outputBytesSlice, []byte(str))
	}
	return outputBytesSlice
}
