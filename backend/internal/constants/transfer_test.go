package constants

import "testing"

func TestCanTransfer(t *testing.T) {
	tests := []struct {
		name string
		from TransferStatus
		to   TransferStatus
		want bool
	}{
		{"pending to confirmed", TransferPending, TransferConfirmed, true},
		{"pending to cancelled", TransferPending, TransferCancelled, true},
		{"confirmed to shipped", TransferConfirmed, TransferShipped, true},
		{"shipped to received", TransferShipped, TransferReceived, true},
		{"received to cancelled", TransferReceived, TransferCancelled, false},
		{"pending to received", TransferPending, TransferReceived, false},
		{"cancelled to shipped", TransferCancelled, TransferShipped, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransfer(tt.from, tt.to); got != tt.want {
				t.Errorf("CanTransfer(%s,%s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestUserRoleValid(t *testing.T) {
	tests := []struct {
		role UserRole
		want bool
	}{
		{RoleHQ, true},
		{RoleStoreManager, true},
		{RoleAdmin, true},
		{"manager", false},
	}
	for _, tt := range tests {
		if got := tt.role.Valid(); got != tt.want {
			t.Errorf("UserRole(%s).Valid() = %v, want %v", tt.role, got, tt.want)
		}
	}
}
