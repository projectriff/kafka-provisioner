package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"net/http"
	"os"
	"strings"
)

type TopicCreationHandler struct {
	Broker  string
	Gateway string
}

func (tch *TopicCreationHandler) HandleTopicCreationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		tch.handlePut(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (tch *TopicCreationHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "URLs should be of the form /<namespace>/<stream-name>\n")
		return
	}

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.ClientID = "kafka-provisioner"
	admin, err := sarama.NewClusterAdmin([]string{tch.Broker}, config)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(os.Stderr, "Error connecting to Kafka broker %q: %v\n", tch.Broker, err)
		_, _ = fmt.Fprintf(w, "Error connecting to Kafka broker %q: %v\n", tch.Broker, err)
		return
	} else {
		defer func() {
			if err := admin.Close(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error disconnecting from Kafka broker %q: %v\n", tch.Broker, err)
			}
		}()
	}

	// NOTE: choice of underscore as separator is important as it is not allowed in k8s names
	topicName := fmt.Sprintf("%s_%s", parts[0], parts[1])
	topicDetail := sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}
	if metadata, err := admin.DescribeTopics([]string{topicName}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(os.Stderr, "Error trying to list topics to see if %q exists: %v\n", topicName, err)
		_, _ = fmt.Fprintf(w, "Error trying to list topics to see if %q exists: %v\n", topicName, err)
		return
	} else if metadata[0].Err == sarama.ErrUnknownTopicOrPartition {
		if err := admin.CreateTopic(topicName, &topicDetail, false); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(os.Stderr, "Error creating topic %q: %v\n", topicName, err)
			_, _ = fmt.Fprintf(w, "Error creating topic %q: %v\n", topicName, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else if metadata[0].Err == sarama.ErrNoError {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(os.Stderr, "Error creating topic %q: %v\n", topicName, err)
		_, _ = fmt.Fprintf(w, "Error creating topic %q: %v\n", topicName, err)
		return
	}

	// Either created or already existed
	w.Header().Set("Content-Type", "application/json")
	res := result{Gateway: tch.Gateway, Topic: topicName}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to write json response: %v", err)
		return
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Reported successful topic %q\n", topicName)
	}
}

type result struct {
	Gateway string `json:"gateway"`
	Topic   string `json:"topic"`
}
