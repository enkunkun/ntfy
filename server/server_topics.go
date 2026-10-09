package server

import (
	"cmp"
	"errors"
	"net/http"
	"slices"

	"heckel.io/ntfy/v2/user"
)

// apiTopicsResponse is the response for GET /v1/topics
type apiTopicsResponse struct {
	Topics []*apiTopicsResponseTopic `json:"topics"`
}

// apiTopicsResponseTopic is a single topic in the GET /v1/topics response
type apiTopicsResponseTopic struct {
	Topic    string `json:"topic"`
	LastTime int64  `json:"last_time"`
}

// handleTopicsGet lists the topics with cached messages that the requesting user is allowed to read,
// most recently active topic first
func (s *Server) handleTopicsGet(w http.ResponseWriter, _ *http.Request, v *visitor) error {
	cached, err := s.messageCache.TopicsWithLastTime()
	if err != nil {
		return err
	}
	topics := make([]*apiTopicsResponseTopic, 0)
	for _, t := range cached {
		if err := s.userManager.Authorize(v.User(), t.Topic, user.PermissionRead); errors.Is(err, user.ErrUnauthorized) {
			continue
		} else if err != nil {
			return err
		}
		topics = append(topics, &apiTopicsResponseTopic{Topic: t.Topic, LastTime: t.LastTime})
	}
	slices.SortFunc(topics, func(a, b *apiTopicsResponseTopic) int {
		return cmp.Or(cmp.Compare(b.LastTime, a.LastTime), cmp.Compare(a.Topic, b.Topic))
	})
	return s.writeJSON(w, &apiTopicsResponse{Topics: topics})
}
