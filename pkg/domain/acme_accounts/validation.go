package acme_accounts

import (
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

// isIdExisting returns an error if not valid, nil if valid
func (service *Service) idExists(id int) bool {
	_, err := service.storage.GetOneAccountById(id)
	return err == nil
}

// nameValid returns true if the specified account name is acceptable and
// false if it is not. This check includes validating specified
// characters and also confirms the name is not already in use by another
// account. If an id is specified, the name will also be accepted if the name
// is already in use by the specified id.
func (service *Service) nameValid(accountName string, accountId *int) bool {
	// basic check
	if !validation.NameValid(accountName) {
		return false
	}

	// make sure the name isn't already in use in storage
	account, err := service.storage.GetOneAccountByName(accountName)
	if err == storage.ErrNoRecord {
		// no rows means name is not in use (valid)
		return true
	} else if err != nil {
		// any other error, invalid
		return false
	}

	// if the returned account is the account being edited, name valid
	if accountId != nil && account.ID == *accountId {
		return true
	}

	return false
}

// GetAvailableAccounts returns a list of accounts that have status == valid
// and have also accepted the ToS (which is probably redundant)
func (service *Service) GetAvailableAccounts() ([]Account, error) {
	accounts, err := service.storage.GetAllAccounts()
	if err != nil {
		return nil, err
	}

	// rewrite accounts in place with only valid accounts
	newIndex := 0
	for i := range accounts {
		if accounts[i].Status == "valid" && accounts[i].AcceptedTos {
			accounts[newIndex] = accounts[i]
			newIndex++
		}
	}
	// truncate accounts
	accounts = accounts[:newIndex]

	return accounts, nil
}
