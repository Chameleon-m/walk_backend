package repository

import (
	"fmt"
	"os"

	"walk_backend/internal/app/model"

	"github.com/gofrs/uuid"
	rabbitmq "github.com/wagslane/go-rabbitmq"
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
	publisher     *rabbitmq.Publisher
	notifyReturn  <-chan rabbitmq.Return
	notifyPublish <-chan rabbitmq.Confirmation
}

var _ PlaceQueueRepositoryInterface = (*PlaceQueueRabbitRepository)(nil)

// NewPlaceQueueRabbitRepository create new queue rabbitmq repository
func NewPlaceQueueRabbitRepository(
	publisher *rabbitmq.Publisher,
	notifyReturn <-chan rabbitmq.Return,
	notifyPublish <-chan rabbitmq.Confirmation,
) *PlaceQueueRabbitRepository {
	return &PlaceQueueRabbitRepository{
		publisher:     publisher,
		notifyReturn:  notifyReturn,
		notifyPublish: notifyPublish,
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
		[]string{os.Getenv("RABBITMQ_ROUTING_PLACE_KEY")},
		rabbitmq.WithPublishOptionsExchange(os.Getenv("RABBITMQ_EXCHANGE_REINDEX")),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsCorrelationID(correlationID.String()),
	); err != nil {
		return err
	}

	for {
		select {
		case ret := <-r.notifyReturn:
			if ret.CorrelationId == correlationID.String() {
				return &NotifyReturnError{ReturnNotify: ret}
			}
			continue
		case confirm := <-r.notifyPublish:
			if !confirm.Ack {
				return &NotifyPublishError{ConfirmNotify: confirm}
			}
			return nil
		}
	}
}
