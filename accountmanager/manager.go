package accountmanager

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go-common-utils/heap"
	"path/filepath"
	"sync"
	"time"
)

const passphrase = "password@rpc"
const IntervalTime = 2 * time.Minute

var AM *AccountManager

func init() {
	AM = NewAccountManager()
}

type AccountManager struct {
	rwmtx    sync.Mutex
	FreeList heap.Heap[*Account]
}

func NewAccountManager() *AccountManager {
	am := &AccountManager{}
	am.FreeList.Init()
	return am
}

// ReadFromFile reads all accounts from the keystore directory
func (am *AccountManager) ReadFromFile() {
	keystoreDir := filepath.Join("keystore")
	ks := keystore.NewKeyStore(keystoreDir, keystore.LightScryptN, keystore.LightScryptP)

	// List all accounts in the keystore
	accounts := ks.Accounts()

	for _, account := range accounts {
		// Assume a common passphrase for simplicity (replace with proper management in a real application)
		err := ks.Unlock(account, passphrase)
		if err != nil {
			fmt.Printf("Failed to unlock account %s: %v\n", account.Address.Hex(), err)
			continue
		}

		// Export and decrypt the private key
		keyJSON, err := ks.Export(account, passphrase, passphrase)
		if err != nil {
			fmt.Printf("Failed to export key for account %s: %v\n", account.Address.Hex(), err)
			continue
		}

		decryptedKey, err := keystore.DecryptKey(keyJSON, passphrase)
		if err != nil {
			fmt.Printf("Failed to decrypt key for account %s: %v\n", account.Address.Hex(), err)
			continue
		}

		// Convert private key to hex string
		privateKeyHex := hex.EncodeToString(crypto.FromECDSA(decryptedKey.PrivateKey))

		// Create and add the account to the FreeList
		am.rwmtx.Lock()
		am.FreeList.PushOne(&Account{
			PrivateKey: privateKeyHex,
			Address:    account.Address,
			UsedTime:   time.Now().Add(-(IntervalTime + 1*time.Minute)),
		})
		am.rwmtx.Unlock()
	}
}

func (am *AccountManager) GetAccount() *Account {
	am.rwmtx.Lock()
	defer am.rwmtx.Unlock()

	one, ok := am.FreeList.PopOne()
	if ok {
		return one
	}

	account, _ := NewAccount()
	return account
}

func (am *AccountManager) PutAccount(a *Account, usedTime time.Time) {
	am.rwmtx.Lock()
	defer am.rwmtx.Unlock()

	if (usedTime != time.Time{}) {
		a.UsedTime = usedTime
	} else {
		a.UsedTime = time.Now()
	}

	am.FreeList.PushOne(a)
}

func (am *AccountManager) GetAccountCount() int {
	am.rwmtx.Lock()
	defer am.rwmtx.Unlock()
	return am.FreeList.Len()
}

type Account struct {
	PrivateKey string
	Address    common.Address
	UsedTime   time.Time
}

func NewAccount() (*Account, error) {
	// Define the keystore directory and create a keystore manager
	keystoreDir := filepath.Join("keystore")
	ks := keystore.NewKeyStore(keystoreDir, keystore.LightScryptN, keystore.LightScryptP)

	// Create a new account with a passphrase (use a secure passphrase in a real application)
	account, err := ks.NewAccount(passphrase)
	if err != nil {
		return nil, errors.New("failed to create new account: " + err.Error())
	}

	// Unlock the account and retrieve the private key
	err = ks.Unlock(account, passphrase)
	if err != nil {
		return nil, errors.New("failed to unlock account: " + err.Error())
	}

	keyJSON, err := ks.Export(account, passphrase, passphrase)
	if err != nil {
		return nil, errors.New("failed to export key: " + err.Error())
	}

	decryptedKey, err := keystore.DecryptKey(keyJSON, passphrase)
	if err != nil {
		return nil, errors.New("failed to decrypt key: " + err.Error())
	}

	// Convert private key to hex string
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(decryptedKey.PrivateKey))

	// Return the Account struct with the private key and address
	return &Account{
		PrivateKey: privateKeyHex,
		Address:    account.Address,
		UsedTime:   time.Now(),
	}, nil
}

func (a *Account) Less(t heap.Element) bool {
	if a.UsedTime.Equal(t.(*Account).UsedTime) {
		return a.Address.Hex() < t.(*Account).Address.Hex()
	}
	return a.UsedTime.Before(t.(*Account).UsedTime)
}

func (a *Account) IsUsable() bool {
	if a.UsedTime.Add(IntervalTime).Before(time.Now()) {
		return true
	}
	return false
}
