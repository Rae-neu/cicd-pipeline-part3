package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(10, 4)
	if result != 6 {
		t.Errorf("Subtract(10, 4) = %d; want 6", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(6, 7)
	if result != 42 {
		t.Errorf("Multiply(6, 7) = %d; want 42", result)
	}
}

func TestDivide(t *testing.T) {
	result, err := Divide(20, 5)

	if err != nil {
		t.Fatalf("Divide(20, 5) returned error: %v", err)
	}

	if result != 4 {
		t.Errorf("Divide(20, 5) = %d; want 4", result)
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(10, 0)

	if err == nil {
		t.Error("Divide(10, 0) should return an error")
	}
}