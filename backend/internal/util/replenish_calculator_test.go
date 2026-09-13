package util

import "testing"

func TestCalculateSuggestQty(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		safetyStock int
		want        int
	}{
		{"above safety", 100, 80, 0},
		{"equal safety", 80, 80, 0},
		{"below safety", 20, 80, 100},
		{"zero stock", 0, 50, 75},
		{"safety zero", 0, 0, 0},
		{"negative safety", 10, -5, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateSuggestQty(tt.quantity, tt.safetyStock); got != tt.want {
				t.Errorf("CalculateSuggestQty(%d,%d) = %d, want %d", tt.quantity, tt.safetyStock, got, tt.want)
			}
		})
	}
}

func TestIsSlowMoving(t *testing.T) {
	tests := []struct {
		name         string
		quantity     int
		monthlySales int
		want         bool
	}{
		{"no stock", 0, 5, false},
		{"slow moving", 500, 10, true},
		{"fast moving", 100, 60, false},
		{"boundary", 100, 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSlowMoving(tt.quantity, tt.monthlySales); got != tt.want {
				t.Errorf("IsSlowMoving(%d,%d) = %v, want %v", tt.quantity, tt.monthlySales, got, tt.want)
			}
		})
	}
}

func TestInventoryStatusLevel(t *testing.T) {
	tests := []struct {
		name     string
		qty      int
		safety   int
		wantText string
	}{
		{"normal", 100, 50, "正常"},
		{"low", 30, 50, "低库存"},
		{"out of stock", 0, 50, "缺货"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InventoryStatusText(tt.qty, tt.safety); got != tt.wantText {
				t.Errorf("InventoryStatusText(%d,%d) = %s, want %s", tt.qty, tt.safety, got, tt.wantText)
			}
		})
	}
}
