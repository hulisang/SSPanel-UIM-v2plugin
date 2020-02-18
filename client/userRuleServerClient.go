package client

import (
	"context"
	"google.golang.org/grpc"
	userruleservice "v2ray.com/core/app/router/command"
)

type UserRuleServerClient struct {
	userruleservice.UserRuleServerClient
}

func NewUserRuleServerClient(client *grpc.ClientConn) *UserRuleServerClient {
	return &UserRuleServerClient{
		UserRuleServerClient: userruleservice.NewUserRuleServerClient(client),
	}
}

func (s *UserRuleServerClient) AddUserRelyRule(targettag string, emails []string) error {
	_, err := s.AddUserRule(context.Background(), &userruleservice.AddUserRuleRequest{
		TargetTag: targettag,
		Email:     emails,
	})
	return err
}

func (s *UserRuleServerClient) RemveUserRelayRule(email []string) error {
	_, err := s.RemoveUserRule(context.Background(), &userruleservice.RemoveUserRequest{
		Email: email,
	})
	return err
}
