package main

import (
	"log"

	"github.com/fm2901/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	/*	account, err := svc.RegisterAccount("+992000000001")
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
	*/
	err := svc.Import("C:/homework/dz17/wallet/data")
	if err != nil {
		log.Print(err)
	}
	err = svc.Export("C:/homework/dz17/wallet/data1")
	if err != nil {
		log.Print(err)
	}

	//svc.ImportFromFile("C:/homework/dz16/wallet/data/accounts.dump")

	//	wallet.CopyFile("C:/homework/dz17/wallet/data/accounts.dump", "C:/homework/dz17/wallet/data/accounts_copy.dump")
}
