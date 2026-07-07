package api

import (
	"aiolimas/db"
	"aiolimas/logging"
	"aiolimas/settings"
	db_types "aiolimas/types"
	"aiolimas/util"
)

func Transact(ctx RequestContext) {
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(ctx.W, 500, "Could not update entry\n%s", err.Error())
		return
	}

	ty := "Purchased"
	if ctx.PP["price"].(float64) < 0 {
		ty = "Sold"
	}

	db.CreateTransaction(
		db_types.Transaction(ty),
		ctx.Uid,
		ctx.PP["id"].(db_types.InfoEntry).ItemId,
		ctx.PP.Get("timezone", us.DefaultTimeZone).(string),
		ctx.PP["price"].(float64),
		ctx.PP["currency"].(string),
	)

	ctx.W.WriteHeader(200)
}

func ListTransactions(ctx RequestContext) {
	item, ok := ctx.PP["id"]
	var id int64 = 0
	if ok {
		id = item.(db_types.InfoEntry).ItemId
	}
	ts, err := db.ListTransactions(ctx.Uid, id)
	if err != nil {
		util.WError(ctx.W, 500, "Could not fetch transaction list\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(200)

	for _, event := range ts {
		j, err := event.ToJson()
		if err != nil {
			logging.ELog(err)
			continue
		}
		ctx.W.Write(j)
		ctx.W.Write([]byte("\n"))
	}
}
