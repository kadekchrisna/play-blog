protoc --proto_path=greet/greetpb/ --go_out=plugins=grpc:greet/greetpb/ greet/greetpb/greet.proto
protoc --proto_path=blog/blogpb/ --go_out=plugins=grpc:blog/blogpb/ blog/blogpb/blog.proto
protoc --proto_path=calculator/calculatorpb/ --go_out=plugins=grpc:calculator/calculatorpb/ calculator/calculatorpb/calculator.proto