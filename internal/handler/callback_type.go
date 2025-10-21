package handler

const (
	CallbackBuy      = "buy"
	CallbackSell     = "sell"
	CallbackStart    = "start"
	CallbackConnect  = "connect"
	CallbackPayment  = "payment"
	CallbackTrial    = "trial"
	CallbackReferral = "referral"
	
	// Multiple subscriptions callbacks
	CallbackMySubscriptions        = "my_subscriptions"
	CallbackOpenSubscription       = "open_subscription"
	CallbackDeactivateSubscription = "deactivate_subscription"
	CallbackRenameSubscription     = "rename_subscription"
	CallbackRenameConfirm         = "rename_confirm"
	
	// Broadcast callbacks
	CallbackBroadcastMenu     = "broadcast_menu"
	CallbackBroadcastToAll    = "broadcast_to_all"
	CallbackBroadcastToAdmins = "broadcast_to_admins"
	CallbackBroadcastConfirm  = "broadcast_confirm"
	CallbackBroadcastCancel   = "broadcast_cancel"
)