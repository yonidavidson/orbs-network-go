package e2e

import (
	"bytes"
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"io/ioutil"
	"net/http"
	"os"
)

type E2EConfig struct {
	Bootstrap   bool
	ApiEndpoint string
}

func getConfig() E2EConfig {
	Bootstrap := len(os.Getenv("API_ENDPOINT")) == 0
	ApiEndpoint := "http://localhost:8080/api/"

	if !Bootstrap {
		ApiEndpoint = os.Getenv("API_ENDPOINT")
	}

	return E2EConfig{
		Bootstrap,
		ApiEndpoint,
	}
}

func SendTransaction(txBuilder *protocol.SignedTransactionBuilder) *services.SendTransactionOutput {
	input := (&client.SendTransactionRequestBuilder{
		SignedTransaction: txBuilder,
	}).Build()

	return &services.SendTransactionOutput{ClientResponse: client.SendTransactionResponseReader(httpPost(input, "send-transaction"))}
}

func CallMethod(txBuilder *protocol.TransactionBuilder) *services.CallMethodOutput {
	input := (&client.CallMethodRequestBuilder{
		Transaction: txBuilder,
	}).Build()

	return &services.CallMethodOutput{ClientResponse: client.CallMethodResponseReader(httpPost(input, "call-method"))}

}

func GetBalance() uint64 {
	m := &protocol.TransactionBuilder{
		ContractName: "BenchmarkToken",
		MethodName:   "getBalance",
	}

	response := CallMethod(m).ClientResponse.OutputArgumentsIterator()
	if response.HasNext() {
		return response.NextOutputArguments().Uint64Value()
	} else {
		return 0
	}
}

func httpPost(input membuffers.Message, method string) []byte {
	res, err := http.Post(getConfig().ApiEndpoint+method, "application/octet-stream", bytes.NewReader(input.Raw()))
	if err != nil {
		panic(err)
	}
	//Expect(err).ToNot(HaveOccurred())
	//Expect(res.StatusCode).To(Equal(http.StatusOK))

	bytes, err := ioutil.ReadAll(res.Body)

	defer res.Body.Close()

	if err != nil {
		panic(err)
	}
	//Expect(err).ToNot(HaveOccurred())

	return bytes
}
