package grpc

import (
	"net"

	"user-service/internal/domain/entity"
	"user-service/internal/domain/service"
	pb "user-service/proto/user/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// toProtoUser converts domain user to proto user
func toProtoUser(user *entity.User) *pb.User {
	protoUser := &pb.User{
		Id:            user.ID.String(),
		Email:         user.Email,
		Username:      user.Username,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		Timezone:      user.Timezone,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
	}

	if user.FirstName != nil {
		protoUser.FirstName = *user.FirstName
	}

	return protoUser
}

// toProtoSession converts domain session to proto session
func toProtoSession(session *entity.Session) *pb.Session {
	protoSession := &pb.Session{
		Id:             session.ID.String(),
		UserId:         session.UserID.String(),
		ExpiresAt:      timestamppb.New(session.ExpiresAt),
		CreatedAt:      timestamppb.New(session.CreatedAt),
		LastActivityAt: timestamppb.New(session.LastActivityAt),
	}

	if session.IPAddress != nil {
		ipStr := session.IPAddress.String()
		protoSession.IpAddress = &ipStr
	}

	if session.UserAgent != nil {
		protoSession.UserAgent = session.UserAgent
	}

	return protoSession
}

// toProtoTokenPair converts service token pair to proto response
func toProtoTokenPair(tokenPair *service.TokenPair) (string, string, *timestamppb.Timestamp, *timestamppb.Timestamp) {
	return tokenPair.AccessToken,
		tokenPair.RefreshToken,
		timestamppb.New(tokenPair.AccessTokenExpiresAt),
		timestamppb.New(tokenPair.RefreshTokenExpiresAt)
}

// parseIPAddress parses IP address string to net.IP
func parseIPAddress(ipStr *string) *net.IP {
	if ipStr == nil || *ipStr == "" {
		return nil
	}
	ip := net.ParseIP(*ipStr)
	if ip == nil {
		return nil
	}
	return &ip
}

// toUserCreate converts proto register request to domain user create
func toUserCreate(req *pb.RegisterRequest) *entity.UserCreate {
	userCreate := &entity.UserCreate{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Timezone: req.Timezone,
	}

	if req.FirstName != "" {
		userCreate.FirstName = &req.FirstName
	}

	return userCreate
}

// toUserUpdate converts proto update request to domain user update
func toUserUpdate(req *pb.UpdateUserRequest) *entity.UserUpdate {
	userUpdate := &entity.UserUpdate{}

	if req.FirstName != nil {
		userUpdate.FirstName = req.FirstName
	}

	if req.Timezone != nil {
		userUpdate.Timezone = req.Timezone
	}

	if req.EmailVerified != nil {
		userUpdate.EmailVerified = req.EmailVerified
	}

	return userUpdate
}
