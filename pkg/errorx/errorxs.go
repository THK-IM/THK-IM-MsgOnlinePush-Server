package errorx

import "github.com/thk-im/thk-im-base-server/errorx"

var (
	ErrParamsError            = errorx.NewErrorX(4000000, "Params Error")
	ErrPermission             = errorx.NewErrorX(4001001, "Permission denied")
	ErrUserNotOnLine          = errorx.NewErrorX(4000002, "User not on line")
	ErrSessionInvalid         = errorx.NewErrorX(4000003, "Invalid session")
	ErrGroupMemberCountBeyond = errorx.NewErrorX(4000004, "group member count beyond")
	ErrGroupAlreadyDeleted    = errorx.NewErrorX(4000005, "group has been deleted")
	ErrSessionType            = errorx.NewErrorX(4000006, "Session type error")
	ErrMessageFormat          = errorx.NewErrorX(4000007, "Message format error")
	ErrMessageTypeNotSupport  = errorx.NewErrorX(4000008, "Message type not support")
	ErrSessionMessageInvalid  = errorx.NewErrorX(4000009, "Invalid session or message")
	ErrGroupAlreadyCreated    = errorx.NewErrorX(4000010, "group has been created")
	ErrSessionMuted           = errorx.NewErrorX(4001001, "Session muted")
	ErrUserMuted              = errorx.NewErrorX(4001002, "User muted")
	ErrUserReject             = errorx.NewErrorX(4001003, "user reject your message")

	ErrServerUnknown         = errorx.NewErrorX(5000000, "Server unknown err")
	ErrServerBusy            = errorx.NewErrorX(5000001, "Server busy")
	ErrMessageDeliveryFailed = errorx.NewErrorX(5004001, "Message delivery failed")
)
