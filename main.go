package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/bobbsley/couponservice/genproto"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultPort = "60000"
)

var validCoupons map[string]int = make(map[string]int)

// readCouponCodes reads coupons from a given .csv file into the memory
func readCouponCodes(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	for i, row := range rows {
		// Skip the header
		if i == 0 {
			continue
		}
		validCoupons[row[0]], err = strconv.Atoi(row[1])
		if err != nil {
			validCoupons[row[0]] = 0
		}
	}
	return nil
}

func init() {
	err := readCouponCodes("coupons.csv")
	if err != nil {
		log.Println("CSV Parse Error. No valid codes found.")
	}
}

func main() {
	port := defaultPort
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	port = fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	svc := &server{}
	pb.RegisterCouponServiceServer(grpcServer, svc)
	healthpb.RegisterHealthServer(grpcServer, svc)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// COPIED FROM shippingservice
// server controls RPC service responses.
type server struct {
	pb.UnimplementedCouponServiceServer
}

// Check is for health checking.
func (s *server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *server) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

// RedeemCoupon verifies whether the coupon is valid and how big the discount is
func (s *server) RedeemCoupon(ctx context.Context, in *pb.CouponRequest) (*pb.CouponResponse, error) {
	var isValid bool
	// Check if the coupon is valid and how big is the discount (in %)
	discount, ok := validCoupons[in.CouponCode]

	if ok {
		isValid = true
	} else {
		isValid = false
		discount = 0
	}

	return &pb.CouponResponse{Validity: isValid, DiscountPercentage: int32(discount)}, nil
}
