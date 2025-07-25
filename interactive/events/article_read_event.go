package events

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Uid int64
	Aid int64
}
