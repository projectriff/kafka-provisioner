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
	"net/http"

	"github.com/Shopify/sarama"
)

var (
	config = sarama.NewConfig()
)

func main() {
	config.Version = sarama.V0_11_0_0
	admin, err := sarama.NewClusterAdmin([]string{"kafka:9092"}, config)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			handlePut(w, r, admin)
		} else {
			w.WriteHeader(404)
		}
	})
	http.ListenAndServe(":8080", nil)
}

func handlePut(w http.ResponseWriter, r *http.Request, admin sarama.ClusterAdmin) {
	topicName := r.URL.Path[1:]
	topicDetail := sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}
	admin.CreateTopic(topicName, &topicDetail, false)
	w.WriteHeader(200)
}
