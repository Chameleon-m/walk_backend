package repository

import (
	"context"
	"os"
	"time"

	"walk_backend/internal/app/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PlaceQueueRabbitRepository place mongodb repo
type PlaceQueueRabbitRepository struct {
	chanel *amqp.Channel
}

func NewPlaceQueueRabbitRepository(chanel *amqp.Channel) *PlaceQueueRabbitRepository {
	return &PlaceQueueRabbitRepository{
		chanel: chanel,
	}
}

func (r *PlaceQueueRabbitRepository) PublishReIndex(id model.ID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.chanel.PublishWithContext(
		ctx,
		os.Getenv("RABBITMQ_EXCHANGE_REINDEX"),
		os.Getenv("RABBITMQ_ROUTING_PLACE_KEY"),
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(id.String()),
		},
	)

	return err
}
