package wallet

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fm2901/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be a greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("balance not enough")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrPaymentExecuted = errors.New("payment already executed")
var ErrFavoriteNotFound = errors.New("favorite not found")

var ColDelimiter = ";"
var RowDelimiter = "\n"

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:    s.nextAccountID,
		Phone: phone}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (acc *types.Account, position int, err error) {
	var account *types.Account
	for pos, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			position = pos
			break
		}
	}

	if account == nil {
		return nil, 0, ErrAccountNotFound
	}

	return account, position, nil
}

func (s *Service) FindPaymentByID(paymentID string) (pay *types.Payment, position int, err error) {
	for pos, payment := range s.payments {
		if payment.ID == paymentID {
			position = pos
			return payment, position, nil
		}
	}
	return nil, 0, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, _, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}

	if payment.Status != types.PaymentStatusInProgress {
		return ErrPaymentExecuted
	}

	account, _, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, _, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	newPayment, err := s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}

	return newPayment, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, _, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (fav *types.Favorite, position int, err error) {
	for pos, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			position = pos
			return favorite, position, nil
		}
	}
	return nil, 0, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, _, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	accountsStr := ""
	for _, account := range s.accounts {
		id := strconv.Itoa(int(account.ID))
		phone := string(account.Phone)
		balance := strconv.Itoa(int(account.Balance))
		accountsStr += id + ColDelimiter + phone + ColDelimiter + balance + RowDelimiter
	}
	accountsStr = accountsStr[:len(accountsStr)-len(RowDelimiter)]

	_, err = file.Write([]byte(accountsStr))
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	content := make([]byte, 0)
	buf := make([]byte, 4)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}

		if err != nil {
			return err
		}
		content = append(content, buf[:read]...)
	}

	data := string(content)
	rows := strings.Split(data, RowDelimiter)
	for _, row := range rows {
		cols := strings.Split(row, ColDelimiter)
		id, _ := strconv.ParseInt(cols[0], 10, 64)
		phone := types.Phone(cols[1])
		balance, _ := strconv.ParseInt(cols[2], 10, 64)
		s.accounts = append(s.accounts, &types.Account{
			ID:      id,
			Phone:   phone,
			Balance: types.Money(balance),
		})
	}
	return nil
}

func CopyFile(from, to string) (err error) {
	src, err := os.Open(from)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := src.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	stats, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	written, err := io.Copy(dst, src)
	if err != nil {
		return err
	}
	if written != stats.Size() {
		return fmt.Errorf("copied size: %d, original size: %d", written, stats.Size())
	}
	return nil
}

func (s *Service) Export(dir string) error {
	if len(s.accounts) > 0 {
		file, err := os.Create(filepath.Join(dir, "accounts.dump"))
		if err != nil {
			log.Print(err)
			return err
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		accountsStr := ""
		for _, account := range s.accounts {
			id := strconv.Itoa(int(account.ID))
			phone := string(account.Phone)
			balance := strconv.Itoa(int(account.Balance))
			accountsStr += id + ColDelimiter + phone + ColDelimiter + balance + RowDelimiter
		}
		accountsStr = accountsStr[:len(accountsStr)-len(RowDelimiter)]

		_, err = file.Write([]byte(accountsStr))
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if len(s.payments) > 0 {
		file, err := os.Create(filepath.Join(dir, "payments.dump"))
		if err != nil {
			log.Print(err)
			return err
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		paymentsStr := ""
		for _, payment := range s.payments {
			id := string(payment.ID)
			accountID := strconv.Itoa(int(payment.AccountID))
			amount := strconv.Itoa(int(payment.Amount))
			category := string(payment.Category)
			status := string(payment.Status)
			paymentsStr += id + ColDelimiter +
				accountID + ColDelimiter +
				amount + ColDelimiter +
				category + ColDelimiter +
				status + RowDelimiter
		}
		paymentsStr = paymentsStr[:len(paymentsStr)-len(RowDelimiter)]

		_, err = file.Write([]byte(paymentsStr))
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if len(s.favorites) > 0 {
		file, err := os.Create(filepath.Join(dir, "favorites.dump"))
		if err != nil {
			log.Print(err)
			return err
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		favoritesStr := ""
		for _, favorite := range s.favorites {
			id := string(favorite.ID)
			accountID := strconv.Itoa(int(favorite.AccountID))
			name := string(favorite.Name)
			amount := strconv.Itoa(int(favorite.Amount))
			category := string(favorite.Category)

			favoritesStr += id + ColDelimiter +
				accountID + ColDelimiter +
				name + ColDelimiter +
				amount + ColDelimiter +
				category + RowDelimiter
		}
		favoritesStr = favoritesStr[:len(favoritesStr)-len(RowDelimiter)]

		_, err = file.Write([]byte(favoritesStr))
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}

func (s *Service) Import(dir string) error {
	file, err := os.Open(dir + "/accounts.dump")
	if err != nil {
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	if err == nil {
		content := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}

			if err != nil {
				return err
			}
			content = append(content, buf[:read]...)
		}

		data := string(content)
		rows := strings.Split(data, RowDelimiter)
		s.nextAccountID = s.accounts[0].ID
		for _, row := range rows {
			cols := strings.Split(row, ColDelimiter)
			id, _ := strconv.ParseInt(cols[0], 10, 64)
			phone := types.Phone(cols[1])
			balance, _ := strconv.ParseInt(cols[2], 10, 64)
			oldAccount, position, err := s.FindAccountByID(id)
			if err != nil {
				return err
			}
			if s.nextAccountID < id {
				s.nextAccountID = id
			}
			curAccount := &types.Account{
				ID:      id,
				Phone:   phone,
				Balance: types.Money(balance),
			}
			if oldAccount != nil {
				s.accounts[position] = curAccount
			} else {
				s.accounts = append(s.accounts, curAccount)
			}
		}
	}

	file, err = os.Open(dir + "/payments.dump")
	if err != nil {
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	if err == nil {
		content := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}

			if err != nil {
				return err
			}
			content = append(content, buf[:read]...)
		}

		data := string(content)
		rows := strings.Split(data, RowDelimiter)
		for _, row := range rows {
			cols := strings.Split(row, ColDelimiter)
			id := string(cols[0])
			accountID, _ := strconv.ParseInt(cols[1], 10, 64)
			amount, _ := strconv.ParseInt(cols[2], 10, 64)
			category := types.PaymentCategory(cols[3])
			status := types.PaymentStatus(cols[4])
			oldPayment, position, err := s.FindPaymentByID(id)
			if err != nil {
				return err
			}
			curPayment := &types.Payment{
				ID:        id,
				AccountID: accountID,
				Amount:    types.Money(amount),
				Category:  category,
				Status:    status,
			}
			if oldPayment != nil {
				s.payments[position] = curPayment
			} else {
				s.payments = append(s.payments, curPayment)
			}
		}
	}


	file, err = os.Open(dir + "/favorites.dump")
	if err != nil {
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

	if err == nil {
		content := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}

			if err != nil {
				return err
			}
			content = append(content, buf[:read]...)
		}

		data := string(content)
		rows := strings.Split(data, RowDelimiter)
		for _, row := range rows {
			cols := strings.Split(row, ColDelimiter)
			id := string(cols[0])
			accountID, _ := strconv.ParseInt(cols[1], 10, 64)
			name := cols[2]
			amount, _ := strconv.ParseInt(cols[3], 10, 64)
			category := types.PaymentCategory(cols[4])
			
			oldFavorite, position, err := s.FindFavoriteByID(id)
			if err != nil {
				return err
			}
			curFavorite := &types.Favorite{
				ID:        id,
				AccountID: accountID,
				Name: name,
				Amount:    types.Money(amount),
				Category:  category,
			}
			if oldFavorite != nil {
				s.favorites[position] = curFavorite
			} else {
				s.favorites = append(s.favorites, curFavorite)
			}
		}
	}

	return nil
}
