package hub

type FeedClient struct {
	Feed       chan []byte
	IsInternal bool
}

func NewFeedClient(isInternal bool) FeedClient {
	return FeedClient{
		Feed:       make(chan []byte),
		IsInternal: isInternal,
	}
}
