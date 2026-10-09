package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"heckel.io/ntfy/v2/message"
	"heckel.io/ntfy/v2/model"
)

func TestStore_TopicsWithLastTime(t *testing.T) {
	forEachBackend(t, func(t *testing.T, s *message.Cache) {
		m1 := model.NewDefaultMessage("mytopic", "message 1")
		m1.Time = 100
		m2 := model.NewDefaultMessage("mytopic", "message 2")
		m2.Time = 300
		m3 := model.NewDefaultMessage("othertopic", "message 3")
		m3.Time = 200
		require.Nil(t, s.AddMessage(m1))
		require.Nil(t, s.AddMessage(m2))
		require.Nil(t, s.AddMessage(m3))

		topics, err := s.TopicsWithLastTime()
		require.Nil(t, err)
		require.ElementsMatch(t, []message.TopicLastTime{
			{Topic: "mytopic", LastTime: 300},
			{Topic: "othertopic", LastTime: 200},
		}, topics)
	})
}

func TestStore_TopicsWithLastTime_ExcludesScheduled(t *testing.T) {
	forEachBackend(t, func(t *testing.T, s *message.Cache) {
		m1 := model.NewDefaultMessage("mytopic", "published")
		m1.Time = 100
		m2 := model.NewDefaultMessage("mytopic", "scheduled")
		m2.Time = time.Now().Add(time.Hour).Unix()
		m3 := model.NewDefaultMessage("othertopic", "scheduled only")
		m3.Time = time.Now().Add(time.Hour).Unix()
		require.Nil(t, s.AddMessage(m1))
		require.Nil(t, s.AddMessage(m2))
		require.Nil(t, s.AddMessage(m3))

		topics, err := s.TopicsWithLastTime()
		require.Nil(t, err)
		require.Equal(t, []message.TopicLastTime{{Topic: "mytopic", LastTime: 100}}, topics)
	})
}

func TestStore_TopicsWithLastTime_IgnoresActionEvents(t *testing.T) {
	forEachBackend(t, func(t *testing.T, s *message.Cache) {
		m1 := model.NewDefaultMessage("mytopic", "message")
		m1.Time = 100
		m2 := model.NewActionMessage(model.MessageDeleteEvent, "mytopic", m1.ID)
		m2.Time = 200
		m3 := model.NewActionMessage(model.MessageClearEvent, "othertopic", "someid")
		m3.Time = 300
		require.Nil(t, s.AddMessage(m1))
		require.Nil(t, s.AddMessage(m2))
		require.Nil(t, s.AddMessage(m3))

		topics, err := s.TopicsWithLastTime()
		require.Nil(t, err)
		require.Equal(t, []message.TopicLastTime{{Topic: "mytopic", LastTime: 100}}, topics)
	})
}

func TestStore_TopicsWithLastTime_Empty(t *testing.T) {
	forEachBackend(t, func(t *testing.T, s *message.Cache) {
		topics, err := s.TopicsWithLastTime()
		require.Nil(t, err)
		require.NotNil(t, topics)
		require.Empty(t, topics)
	})
}

func TestStore_TopicsWithLastTime_NopStore(t *testing.T) {
	s, err := message.NewNopStore()
	require.Nil(t, err)
	t.Cleanup(func() { s.Close() })
	require.Nil(t, s.AddMessage(model.NewDefaultMessage("mytopic", "message")))

	topics, err := s.TopicsWithLastTime()
	require.Nil(t, err)
	require.NotNil(t, topics)
	require.Empty(t, topics)
}
