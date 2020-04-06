run from root of project:

protoc .\calculator\calculatorpb\calculator.proto --go_out=plugins=grpc:.

go run .\calculator\calculator_server\server.go 
go run .\calculator\calculator_client\client.go