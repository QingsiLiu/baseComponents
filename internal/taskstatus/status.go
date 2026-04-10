package taskstatus

const (
	Pending int32 = iota
	Running
	Completed
	Canceled
	Failed
)

// Name 返回状态名称。
func Name(status int32) string {
	switch status {
	case Pending:
		return "pending"
	case Running:
		return "running"
	case Completed:
		return "completed"
	case Canceled:
		return "canceled"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}
