package grpclib

import (
	"context"
	"gim/pkg/gerrors"
	"gim/pkg/logger"
	"gim/pkg/pb"
	"gim/pkg/rpc"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// NewInterceptor 生成GRPC过滤器
func NewInterceptor(name string, whitelistMethod map[string]int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer gerrors.LogPanic(name, ctx, req, info, &err)

		md, _ := metadata.FromIncomingContext(ctx)
		resp, err = handleWithAuth(ctx, req, info, handler, whitelistMethod)
		logger.Logger.Debug(name, zap.Any("method", info.FullMethod), zap.Any("md", md), zap.Any("req", req),
			zap.Any("resp", resp), zap.Error(err))

		s, _ := status.FromError(err)
		if s.Code() != 0 && s.Code() < 1000 {
			md, _ := metadata.FromIncomingContext(ctx)
			logger.Logger.Error(name, zap.String("method", info.FullMethod), zap.Any("md", md), zap.Any("req", req),
				zap.Any("resp", resp), zap.Error(err), zap.String("stack", gerrors.GetErrorStack(s)))
		}
		return
	}
}

// handleWithAuth 处理鉴权逻辑
func handleWithAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, whitelistMethod map[string]int) (interface{}, error) {
	if _, ok := whitelistMethod[info.FullMethod]; !ok {
		userId, deviceId, err := GetCtxData(ctx)
		if err != nil {
			return nil, err
		}
		token, err := GetCtxToken(ctx)
		if err != nil {
			return nil, err
		}

		_, err = rpc.BusinessIntClient.Auth(ctx, &pb.AuthReq{UserId: userId, DeviceId: deviceId, Token: token})
		if err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}
