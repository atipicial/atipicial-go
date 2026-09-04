package rpcsrv

import (
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/rpcevent"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/vmstate"
)

// notificationEventComparator is a comparator for notification events.
type notificationEventComparator struct {
	filter atipicialrpc.SubscriptionFilter
}

// EventID returns the event ID for the notification event comparator.
func (s notificationEventComparator) EventID() atipicialrpc.EventID {
	return atipicialrpc.NotificationEventID
}

// Filter returns the filter for the notification event comparator.
func (c notificationEventComparator) Filter() atipicialrpc.SubscriptionFilter {
	return c.filter
}

// notificationEventContainer is a container for a notification event.
type notificationEventContainer struct {
	ntf *state.ContainedNotificationEvent
}

// EventID returns the event ID for the notification event container.
func (c notificationEventContainer) EventID() atipicialrpc.EventID {
	return atipicialrpc.NotificationEventID
}

// EventPayload returns the payload for the notification event container.
func (c notificationEventContainer) EventPayload() any {
	return c.ntf
}

func processAppExecResults(aers []state.AppExecResult, filter *atipicialrpc.NotificationFilter) []state.ContainedNotificationEvent {
	var notifications []state.ContainedNotificationEvent
	for _, aer := range aers {
		if aer.VMState == vmstate.Halt {
			notifications = append(notifications, filterEvents(aer.Events, aer.Container, filter)...)
		}
	}
	return notifications
}

func filterEvents(events []state.NotificationEvent, container util.Uint256, filter *atipicialrpc.NotificationFilter) []state.ContainedNotificationEvent {
	var notifications []state.ContainedNotificationEvent
	for _, evt := range events {
		ntf := state.ContainedNotificationEvent{
			Container:         container,
			NotificationEvent: evt,
		}
		if filter == nil || rpcevent.Matches(&notificationEventComparator{
			filter: *filter,
		}, &notificationEventContainer{ntf: &ntf}) {
			notifications = append(notifications, ntf)
		}
	}
	return notifications
}
