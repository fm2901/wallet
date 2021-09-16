package wallet

import (
	"testing"
	"github.com/fm2901/wallet/pkg/types"
)


func TestService_RegisterAccount_success(t *testing.T) {
	svc := Service{}
	_, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_RegisterAccount_accountExists(t *testing.T) {
	svc := Service{}
	_, err := svc.RegisterAccount("+992000000001")
	_, err = svc.RegisterAccount("+992000000001")
	if err != ErrPhoneRegistered {
		t.Error("Не сработала проверка дублирования номера телефона")
		return
	}
}


func TestService_Deposit_success(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")
	depositSum := types.Money(100)
	err = svc.Deposit(account.ID, depositSum)
	
	if err != nil {
		t.Error(err)
		return
	}

	if account.Balance != depositSum {
		t.Errorf("invalid result, expected: %v, actual: %v", depositSum, account.Balance)
		return
	}
}


func TestService_Pay_success(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")
	
	depositSum := types.Money(100)
	err = svc.Deposit(account.ID, depositSum)
	
	if err != nil {
		t.Error(err)
		return
	}

	paySum := types.Money(50)
	payment, err := svc.Pay(account.ID, paySum, "auto")

	if payment.Amount != paySum {
		t.Errorf("invalid result, expected: %v, actual: %v", paySum, payment.Amount)
		return
	}
}
/*
payments := []types.Payment{
		{ID: 1, Category: "auto", Amount: 1_000},
		{ID: 2, Category: "auto", Amount: 2_000},
		{ID: 3, Category: "home", Amount: 3_000},
		{ID: 4, Category: "fun", Amount: 4_000},
	}
	expected := map[types.Category]types.Money{
		"auto" : 1500,
		"home" : 3000,
		"fun"  : 4000,
	}
	result := CategoriesAvg(payments)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
*/