package main

import (
	"testing"

	"golang.org/x/net/context"

	pb "couponservice/genproto"
)

func TestValidCoupon(t *testing.T) {
	s := server{}

	req := &pb.CouponRequest{CouponCode: "newsletter15"}

	res, err := s.RedeemCoupon(context.Background(), req)
	if err != nil {
		t.Errorf("TestValidCoupon (%v) failed", err)
	}
	if res.Validity != true || res.DiscountPercentage != 15 {
		t.Errorf("TestValidCoupon: Validity '%v', DiscountPercentage '%v' does not match expected '%v', '%v", res.Validity, res.DiscountPercentage, true, 15)
	}
}

func TestInvalidCoupon(t *testing.T) {
	s := server{}

	req := &pb.CouponRequest{CouponCode: "notindb"}

	res, err := s.RedeemCoupon(context.Background(), req)
	if err != nil {
		t.Errorf("TestValidCoupon (%v) failed", err)
	}
	if res.Validity != false || res.DiscountPercentage != 0 {
		t.Errorf("TestValidCoupon: Validity '%v', DiscountPercentage '%v' does not match expected '%v', '%v", res.Validity, res.DiscountPercentage, false, 0)
	}
}
