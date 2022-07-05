package bot

import (
	botDB "tbot/pkg/db"
	"tbot/pkg/db/memdb"
	"testing"
)

func Test_buildMessages(t *testing.T) {

	t.Run("count_messages", func(t *testing.T) {

		// one record == one message
		res := buildMessages(memdb.MockPurchase)

		if len(res) != 1 {
			t.Fatalf("buildMessages() got len = %d, want len = %d", len(res), 1)
		}

		// different query type must result in different messages
		purchAuction := memdb.MockPurchase
		purchAuction.QueryType = botDB.TodayAuction
		purchGo := memdb.MockPurchase
		purchGo.QueryType = botDB.TodayGo

		res = buildMessages(purchAuction, purchGo)

		if len(res) != 2 {
			t.Fatalf("buildMessages() got len = %d, want len = %d", len(res), 2)
		}

		// same query type must result in common message
		purchAuctionAgain := memdb.MockPurchase
		purchAuctionAgain.QueryType = botDB.TodayAuction

		res = buildMessages(purchAuction, purchAuctionAgain)

		if len(res) != 1 {
			t.Fatalf("buildMessages() got len = %d, want len = %d", len(res), 1)
		}
	})

	t.Run("markdown_escaping", func(t *testing.T) {

		purchAuction := memdb.MockPurchase
		purchAuction.QueryType = botDB.TodayAuction
		purchGo := memdb.MockPurchase
		purchGo.QueryType = botDB.TodayGo
		purchFuture := memdb.MockPurchase
		purchFuture.QueryType = botDB.FutureAuction
		res := buildMessages(memdb.MockPurchase, purchAuction, purchGo, purchFuture)

		for i := range res {
			slash := 0
			// seen := false
			for idx, r := range res[i] {
				switch r {
				case '\\': // one slash
					slash++
				case '[', ']', '(', ')', '{', '}', '~',
					'`', '>', '#', '+', '-', '=', '|', '.', '!':
					if slash != 1 {
						t.Errorf("buildMessages(): unescaped char '%c' at position %d in message '%s'",
							r, idx, res[i])
					}
					slash = 0
				default:
					slash = 0
				}
			}
		}
	})
}
