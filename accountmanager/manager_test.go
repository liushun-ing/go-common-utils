package accountmanager

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

func TestNewAccount(t *testing.T) {
	Convey("create 2 accouts", t, func() {
		os.RemoveAll("keystore")

		AM.ReadFromFile()

		account1 := AM.GetAccount()
		So(AM.GetAccountCount(), ShouldEqual, 0)

		account2 := AM.GetAccount()
		fmt.Println(account2)
		_, err := crypto.HexToECDSA(account2.PrivateKey)
		So(err, ShouldBeNil)
		So(AM.GetAccountCount(), ShouldEqual, 0)

		AM.PutAccount(account1, time.Time{})
		So(AM.GetAccountCount(), ShouldEqual, 1)
		AM.PutAccount(account2, time.Time{})
		So(AM.GetAccountCount(), ShouldEqual, 2)
	})
}
