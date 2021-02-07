package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/bobbsley/couponservice/genproto"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultPort = "60000"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

var validCoupons map[string]int = make(map[string]int)

// readCouponCodes reads coupons from a given .csv file into the memory.
// Returns error when CSV file doesn't exist, isn't parsable or has no discount codes.
func readCouponCodes(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		log.Error("CSV file not found")
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		log.Error("error reading the CSV file")
		return err
	}
	for i, row := range rows {
		// Skip the header
		if i == 0 {
			continue
		}
		discount, err := strconv.Atoi(row[1])
		if err != nil {
			log.Warnf("coupn code \"%v\" with incorrect value %v found in CSV file", row[1])
			discount = 0
		} else if discount < 0 {
			log.Warnf("coupon code \"%v\" with discount %v%% smaller than 0% found in CSV file", row[1])
			discount = 0
		} else if discount > 100 {
			log.Warnf("coupon code \"%v\" with discount %v%% bigger than 100% found in CSV file", row[1])
			discount = 100
		}
		validCoupons[row[0]] = discount
		log.Infof("coupon code \"%v\" with discount %v%% added", row[0], discount)
	}
	if len(validCoupons) == 0 {
		return err
	}
	return nil
}

func init() {
	err := readCouponCodes("coupons.csv")
	if err != nil {
		log.Infof("no discount codes found")
	}
}

func main() {
	port := defaultPort
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	port = fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	svc := &server{}
	pb.RegisterCouponServiceServer(grpcServer, svc)
	healthpb.RegisterHealthServer(grpcServer, svc)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server controls RPC service responses.
type server struct {
	pb.UnimplementedCouponServiceServer
}

// COPIED FROM shippingservice
// Check is for health checking.
func (s *server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

// COPIED FROM shippingservice
func (s *server) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

// RedeemCoupon verifies whether the coupon is valid and how big the discount is
func (s *server) RedeemCoupon(ctx context.Context, in *pb.CouponRequest) (*pb.CouponResponse, error) {
	coupon := in.CouponCode
	log.Infof("verifying coupon %s", coupon)
	var isValid bool
	// Check if the coupon is valid and how big is the discount (in %)
	discount, ok := validCoupons[coupon]

	if ok {
		isValid = true
		log.Infof("coupon %s valid with %v%% discount", coupon, discount)
	} else {
		isValid = false
		discount = 0
		log.Infof("coupon %s invalid", coupon)
	}

	return &pb.CouponResponse{Validity: isValid, DiscountPercentage: int32(discount)}, nil
}
