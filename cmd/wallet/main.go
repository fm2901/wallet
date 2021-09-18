package main

import (
	"fmt"
	"os"
	"github.com/fm2901/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = svc.Deposit(account.ID, 100)
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")
		}
		return
	}
	payment, err := svc.Pay(account.ID, 50, "auto")
	svc.FavoritePayment(payment.ID, "myFavorite")

	svc.Export("C:/homework/dz17/wallet/data")
	svc.Import("C:/homework/dz17/wallet/data")

	file, err := os.Open("C:/homework/dz17/wallet/data/favorites.dump")
	if file == nil {
		fmt.Println("file is nill")
	}
	fmt.Println(file)
	//svc.ImportFromFile("C:/homework/dz16/wallet/data/accounts.dump")

	//	wallet.CopyFile("C:/homework/dz17/wallet/data/accounts.dump", "C:/homework/dz17/wallet/data/accounts_copy.dump")
}
