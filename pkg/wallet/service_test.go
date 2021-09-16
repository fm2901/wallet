package wallet

import (
	"testing"
	"github.com/fm2901/wallet/pkg/types"
	"reflect"
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

func TestService_FindAccountByID_success(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")

	if err != nil {
		t.Error(err)
	}
	
	result, err := svc.FindAccountByID(account.ID)

	if !reflect.DeepEqual(account, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", account, result)
	}
}

func TestService_FindAccountByID_notFound(t *testing.T) {
	svc := Service{}
	_, err := svc.RegisterAccount("+992000000001")

	if err != nil {
		t.Error(err)
	}
	
	result, err := svc.FindAccountByID(0)

	if err != ErrAccountNotFound {
		t.Errorf("invalid result, expected: %v, actual: %v", ErrAccountNotFound, result)
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := Service{}
	
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Error(err)
	}
	
	err = svc.Deposit(account.ID, 100)
	if err != nil {
		t.Error(err)
	}

	payment, err := svc.Pay(account.ID, 50, "auto")
	if err != nil {
		t.Error(err)
	}

	err = svc.Reject(payment.ID)
	if err != nil {
		t.Error(err)
	}

	if payment.Status != types.PaymentStatusFail {
		t.Errorf("invalid result, expected: %v, actual: %v", types.PaymentStatusFail, payment.Status)
	}
}

func TestService_Reject_notFound(t *testing.T) {
	svc := Service{}
	
	err := svc.Reject("123")
	if err != ErrPaymentNotFound {
		t.Errorf("invalid result, expected: %v, actual: %v", ErrPaymentNotFound, err)
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