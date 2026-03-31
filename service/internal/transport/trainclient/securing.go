package trainclient

import (
	"context"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	trainpb "github.com/mihnpro/UserServiceProtos/gen/go/train_service_api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// SecuringTrainClient overwrites user_id on each RPC from JWT context (except Health).
type SecuringTrainClient struct {
	inner trainpb.TrainServiceClient
}

func NewSecuringTrainClient(inner trainpb.TrainServiceClient) *SecuringTrainClient {
	return &SecuringTrainClient{inner: inner}
}

func (c *SecuringTrainClient) Health(ctx context.Context, in *trainpb.HealthRequest, opts ...grpc.CallOption) (*trainpb.HealthResponse, error) {
	return c.inner.Health(ctx, in, opts...)
}

func (c *SecuringTrainClient) GetTrainingPlans(ctx context.Context, in *trainpb.GetTrainingPlansRequest, opts ...grpc.CallOption) (*trainpb.GetTrainingPlansResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.GetTrainingPlansRequest)
	out.UserId = uid
	return c.inner.GetTrainingPlans(ctx, out, opts...)
}

func (c *SecuringTrainClient) CreateTrainingPlan(ctx context.Context, in *trainpb.CreateTrainingPlanRequest, opts ...grpc.CallOption) (*trainpb.CreateTrainingPlanResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.CreateTrainingPlanRequest)
	out.UserId = uid
	return c.inner.CreateTrainingPlan(ctx, out, opts...)
}

func (c *SecuringTrainClient) GetTrainingPlan(ctx context.Context, in *trainpb.GetTrainingPlanRequest, opts ...grpc.CallOption) (*trainpb.GetTrainingPlanResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.GetTrainingPlanRequest)
	out.UserId = uid
	return c.inner.GetTrainingPlan(ctx, out, opts...)
}

func (c *SecuringTrainClient) UpdateTrainingPlan(ctx context.Context, in *trainpb.UpdateTrainingPlanRequest, opts ...grpc.CallOption) (*trainpb.UpdateTrainingPlanResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.UpdateTrainingPlanRequest)
	out.UserId = uid
	return c.inner.UpdateTrainingPlan(ctx, out, opts...)
}

func (c *SecuringTrainClient) DeleteTrainingPlan(ctx context.Context, in *trainpb.DeleteTrainingPlanRequest, opts ...grpc.CallOption) (*trainpb.DeleteTrainingPlanResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.DeleteTrainingPlanRequest)
	out.UserId = uid
	return c.inner.DeleteTrainingPlan(ctx, out, opts...)
}

func (c *SecuringTrainClient) AdjustTrainingPlan(ctx context.Context, in *trainpb.AdjustTrainingPlanRequest, opts ...grpc.CallOption) (*trainpb.AdjustTrainingPlanResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.AdjustTrainingPlanRequest)
	out.UserId = uid
	return c.inner.AdjustTrainingPlan(ctx, out, opts...)
}

func (c *SecuringTrainClient) GetNextWorkout(ctx context.Context, in *trainpb.GetNextWorkoutRequest, opts ...grpc.CallOption) (*trainpb.GetNextWorkoutResponse, error) {
	uid, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out := proto.Clone(in).(*trainpb.GetNextWorkoutRequest)
	out.UserId = uid
	return c.inner.GetNextWorkout(ctx, out, opts...)
}

func requireUserID(ctx context.Context) (string, error) {
	u, err := customContext.GetUser(ctx)
	if err != nil || u == nil {
		return "", status.Error(codes.Unauthenticated, "authentication required")
	}
	return u.UserID.String(), nil
}
