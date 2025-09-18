package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "VKR_gateway_service/gen/go" // путь к сгенерированным файлам

	"google.golang.org/grpc"
)

func main() {
	// Подключаемся к серверу
	conn, err := grpc.Dial("localhost:5104", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	// Создаем gRPC-клиента
	client := pb.NewSemanticServiceClient(conn)

	// Отправляем запрос
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req := &pb.SearchRequest{InputData: "Machine learning for scientific research"}
	res, err := client.SearchPaper(ctx, req)
	if err != nil {
		log.Fatalf("Ошибка при вызове SearchPaper: %v", err)
	}

	fmt.Println("Ответ сервера:", res)
}
