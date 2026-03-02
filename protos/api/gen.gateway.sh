protoc --proto_path=. \
  --go_out=../gen/go/userServicepb/gateway \
  --go_opt=paths=source_relative \
  --go-grpc_out=../gen/go/userServicepb/gateway \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=../gen/go/userServicepb/gateway \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  --openapiv2_out=../swagger \
  --openapiv2_opt=logtostderr=true \
  userServiceGateway.proto