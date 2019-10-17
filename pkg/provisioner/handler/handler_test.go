package handler_test

import (
	"fmt"
	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/kafka-provisioner/pkg/provisioner/handler"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Provisioner", func() {

	const (
		existingTopicNamespace = "some-namespace"
		existingTopicName      = "some-topic"
	)

	var (
		broker              *sarama.MockBroker
		creationHandlerFunc *handler.TopicCreationHandler
		responseRecorder    *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		t := GinkgoT()
		brokerId := int32(1)
		By("Given a test broker")
		broker = sarama.NewMockBroker(t, brokerId)

		existingTopic := fmt.Sprintf("%s_%s", existingTopicNamespace, existingTopicName)
		By(fmt.Sprintf("With the existing topic %s", existingTopic))
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetController(broker.BrokerID()).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader(existingTopic, 0, broker.BrokerID()),
		})

		By("Given the provisioner pointing to the test broker")
		creationHandlerFunc = &handler.TopicCreationHandler{
			Gateway: "liiklus.example.com",
			Broker:  broker.Addr(),
		}
		responseRecorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		broker.Close()
	})

	Context("provisions topics on demand", func() {
		It("returns 200 if the topic already exists", func() {
			path := fmt.Sprintf("/%s/%s", existingTopicNamespace, existingTopicName)
			request := httptest.NewRequest("PUT", path, nil)

			http.HandlerFunc(creationHandlerFunc.HandleTopicCreationRequest).ServeHTTP(responseRecorder, request)

			expectedCode := http.StatusOK
			actualCode := responseRecorder.Code
			Expect(actualCode).To(Equal(expectedCode),
				fmt.Sprintf("Expected %d after topic creation request but got %d", expectedCode, actualCode))
		})
	})
})
