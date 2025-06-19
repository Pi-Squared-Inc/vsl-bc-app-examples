package ethrpc

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/rpc"
)

func CallContextWithJSONResponse(client *rpc.Client, ctx context.Context, method string, args ...interface{}) (*string, error) {
	var result interface{}
	err := client.CallContext(ctx, &result, method, args...)
	if err != nil {
		return nil, err
	}
	resultJson, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	resultJsonString := string(resultJson)
	return &resultJsonString, nil
}
