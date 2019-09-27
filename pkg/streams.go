package pkg

import (
	"errors"
	"fmt"
	"sort"
)

type Stream struct {
	Name          string
	Messages      map[string]StreamMessage
}

func (s *Stream) AddMessage(id string, message map[string]interface{}) StreamMessage {
	if s.Messages == nil {
		s.Messages = make(map[string]StreamMessage)
	}

	streamMessage := StreamMessage{
		ID:      id,
		Content: message,
	}

	s.Messages[id] = streamMessage

	return streamMessage
}

func (s *Stream) GetMessage(ID string) (*StreamMessage, error) {
	m, ok := s.Messages[ID]
	if !ok {
		return nil, errors.New("there are no messages with given ID")
	}

	return &m, nil
}

func (s *Stream) GetMessagesList() []string {
	var list []string
	for _, m := range s.Messages {
		list = append(list, m.ID)
	}

	sort.Strings(list)

	return list
}

func (s *Stream) MessagesCount() int {
	return len(s.Messages)
}

type StreamMessage struct {
	ID      string
	Content map[string]interface{}
}

func (m *StreamMessage) ParseContent() string {
	var list []string
	for k, _ := range m.Content {
		list = append(list, k)
	}

	sort.Strings(list)

	var content string
	for _, i := range list {
		content += fmt.Sprintf("Field: %s\r\n", i)
		content += fmt.Sprintf("Value: %s\r\n\r\n", m.Content[i])
	}

	return content
}

type Streams struct {
	Collection map[string]*Stream
}

func (s *Streams) Push(stream *Stream) {
	if s.Collection == nil {
		s.Collection = make(map[string]*Stream)
	}

	if _, ok := s.Collection[stream.Name]; ok {
		return
	}

	s.Collection[stream.Name] = stream
}

func (s *Streams) Find(key string) *Stream {
	stream, ok := s.Collection[key]

	if !ok {
		return nil
	}

	return stream
}