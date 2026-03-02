protoc --proto_path=. \
  --go_out=../gen/go/userServicepb \
  --go_opt=paths=source_relative \
  --go-grpc_out=../gen/go/userServicepb \
  --go-grpc_opt=paths=source_relative \
    userService.proto