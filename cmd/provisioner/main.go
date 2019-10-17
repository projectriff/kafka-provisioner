/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/projectriff/kafka-provisioner/pkg/provisioner/handler"
	"log"
	"net/http"
	"os"
)

var (
	gateway = os.Getenv("GATEWAY")
	broker  = os.Getenv("BROKER")
)

func main() {
	if gateway == "" {
		log.Fatal("Environment variable GATEWAY should contain the host and port of a liiklus gRPC endpoint")
	}
	if broker == "" {
		log.Fatal("Environment variable BROKER should contain the host and port of a Kafka broker")
	}
	creationHandler := &handler.TopicCreationHandler{
		Gateway: gateway,
		Broker:  broker,
	}
	http.HandleFunc("/", creationHandler.HandleTopicCreationRequest)
	http.ListenAndServe(":8080", nil)
}
