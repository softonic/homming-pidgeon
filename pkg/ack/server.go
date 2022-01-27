package ack

import (
	"github.com/google/uuid"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"k8s.io/klog"
	"sync"
)

type Server struct {
	clients      map[string]proto.AckService_GetMessagesServer
	InputChannel <-chan messages.Ack
	mu           sync.RWMutex
}

func (s *Server) addClient(uid string, client proto.AckService_GetMessagesServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[uid] = client
	klog.V(1).Info("New client connected")
}

func (s *Server) getClients() []proto.AckService_GetMessagesServer {
	var clients []proto.AckService_GetMessagesServer

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *Server) GetMessages(req *proto.EmptyRequest, client proto.AckService_GetMessagesServer) error {
	uid := uuid.Must(uuid.NewRandom()).String()

	s.addClient(uid, client)

	for message := range s.InputChannel {
		klog.V(1).Info("Sending ACK to clients")
		for _, client := range s.getClients() {
			client.Send(&proto.Message{
				Body: message.Body,
				Ack:  message.Ack,
			})
		}
	}
	return nil
}
