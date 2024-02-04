package main

import (
	"context"
	"encoding/json"
	"go-api-payments-ecommerce/internal/entity"
	"go-api-payments-ecommerce/pkg/rabbitmq"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// {"order_id": "IWYDUYWSAS", "card_hash": "card_hash", "total": 500.00}
func main() {
	ctx := context.Background()
	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	msgs := make(chan amqp.Delivery)
	go rabbitmq.Consume(ch, msgs, "orders") // rodando em background

	for msg := range msgs {
		var orderRequest entity.OrderRequest
		err := json.Unmarshal(msg.Body, &orderRequest)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		response, err := orderRequest.Process()
		if err != nil {
			slog.Error(err.Error())
			break
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			slog.Error(err.Error())
			break
		}

		err = rabbitmq.Publish(ctx, ch, string(responseJSON), "amq.direct")
		if err != nil {
			slog.Error(err.Error())
			break
		}
		msg.Ack(false) //confirmação de recebimento da mensagem
		slog.Info("Order processed")
	}
}
