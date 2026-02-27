package validator

import (
	"testing"
)

func TestValidateBDPhone(t *testing.T) {
	type Phone struct {
		Number string `validate:"required,phone_bd"`
	}

	valid := []string{
		"+8801712345678",
		"+8801912345678",
		"+8801312345678",
		"01712345678",
		"01912345678",
	}

	invalid := []string{
		"1234567890",
		"+8800112345678",
		"01012345678",
		"+880171234567",   // too short
		"+88017123456789", // too long
		"",
	}

	for _, num := range valid {
		err := Validate.Struct(Phone{Number: num})
		if err != nil {
			t.Errorf("expected %q to be valid BD phone, got error: %v", num, err)
		}
	}

	for _, num := range invalid {
		err := Validate.Struct(Phone{Number: num})
		if err == nil {
			t.Errorf("expected %q to be invalid BD phone", num)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	type ID struct {
		Value string `validate:"required,uuid"`
	}

	valid := "550e8400-e29b-41d4-a716-446655440000"
	err := Validate.Struct(ID{Value: valid})
	if err != nil {
		t.Errorf("expected valid UUID, got error: %v", err)
	}

	invalid := "not-a-uuid"
	err = Validate.Struct(ID{Value: invalid})
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestValidateDecimal(t *testing.T) {
	type Price struct {
		Amount string `validate:"required,decimal"`
	}

	valid := []string{"150.00", "0.5", "1000", "99.9999"}
	for _, v := range valid {
		err := Validate.Struct(Price{Amount: v})
		if err != nil {
			t.Errorf("expected %q to be valid decimal, got error: %v", v, err)
		}
	}

	invalid := []string{"abc", "12.345.6", "-5.00", ""}
	for _, v := range invalid {
		err := Validate.Struct(Price{Amount: v})
		if err == nil {
			t.Errorf("expected %q to be invalid decimal", v)
		}
	}
}

func TestFormatErrors(t *testing.T) {
	type Input struct {
		Email string `validate:"required,email"`
		Phone string `validate:"required,phone_bd"`
	}

	err := Validate.Struct(Input{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	formatted := FormatErrors(err)
	if len(formatted) == 0 {
		t.Error("expected formatted errors to be non-empty")
	}

	if _, ok := formatted["Email"]; !ok {
		t.Error("expected Email field in formatted errors")
	}
}
