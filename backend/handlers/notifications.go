package handlers

import "github.com/tonk/coworker/services"

var notifSvc *services.NotificationService

// InitNotifications stores the notification service reference for use by handlers.
func InitNotifications(svc *services.NotificationService) {
	notifSvc = svc
}
