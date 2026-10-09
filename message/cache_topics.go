package message

// TopicLastTime is a topic name and the time of the most recent published message in it
type TopicLastTime struct {
	Topic    string
	LastTime int64
}

// TopicsWithLastTime returns all topics with published messages in the cache, together with the
// time of the most recent message. Scheduled messages that are not yet published and action events
// (e.g. message_delete, message_clear) are not considered.
func (c *Cache) TopicsWithLastTime() ([]TopicLastTime, error) {
	rows, err := c.db.ReadOnly().Query(c.queries.selectTopicsWithLastTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	topics := make([]TopicLastTime, 0)
	for rows.Next() {
		var t TopicLastTime
		if err := rows.Scan(&t.Topic, &t.LastTime); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return topics, nil
}
