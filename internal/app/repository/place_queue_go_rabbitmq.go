package repository

import (
	"fmt"

	"walk_backend/internal/app/model"

	"github.com/gofrs/uuid"
	rabbitmq "github.com/wagslane/go-rabbitmq"
	"golang.org/x/net/context"
)

// NotifyReturnError ...
type NotifyReturnError struct {
	ReturnNotify rabbitmq.Return
}

// Error ...
func (e *NotifyReturnError) Error() string {
	return fmt.Sprintf(
		"publish %s return ReplyCode %d, ReplyText %s, Exchange %s, RoutingKey %s",
		e.ReturnNotify.CorrelationId,
		e.ReturnNotify.ReplyCode,
		e.ReturnNotify.ReplyText,
		e.ReturnNotify.Exchange,
		e.ReturnNotify.RoutingKey,
	)
}

// NotifyPublishError ...
type NotifyPublishError struct {
	ConfirmNotify rabbitmq.Confirmation
}

// Error ...
func (e *NotifyPublishError) Error() string {
	return fmt.Sprintf(
		"publish %d is not ack, reconnection: %d",
		e.ConfirmNotify.DeliveryTag,
		e.ConfirmNotify.ReconnectionCount,
	)
}

// PlaceQueueRabbitRepository place queue repository
type PlaceQueueRabbitRepository struct {
	ctx               context.Context
	publisher         *rabbitmq.Publisher
	reindexExchange   string
	reindexRoutingKey string
}

// NewPlaceQueueRabbitRepository create new queue rabbitmq repository
func NewPlaceQueueRabbitRepository(
	ctx context.Context,
	publisher *rabbitmq.Publisher,
	reindexExchange string,
	reindexRoutingKey string,
) *PlaceQueueRabbitRepository {
	return &PlaceQueueRabbitRepository{
		ctx:               ctx,
		publisher:         publisher,
		reindexExchange:   reindexExchange,
		reindexRoutingKey: reindexRoutingKey,
	}
}

// PublishReIndex publish re index
func (r *PlaceQueueRabbitRepository) PublishReIndex(id model.ID) error {

	correlationID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	if err := r.publisher.Publish(
		[]byte(id.String()),
		[]string{r.reindexRoutingKey},
		rabbitmq.WithPublishOptionsExchange(r.reindexExchange),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsCorrelationID(correlationID.String()),
	); err != nil {
		return err
	}

	for {
		if err := r.ctx.Err(); err != nil {
			return err
		}

		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case ret := <-r.publisher.NotifyReturn():
			if ret.CorrelationId == correlationID.String() {
				return &NotifyReturnError{ReturnNotify: ret}
			}
			continue
		case confirm := <-r.publisher.NotifyPublish():
			if !confirm.Ack {
				return &NotifyPublishError{ConfirmNotify: confirm}
			}
			return nil
		}
	}
}
