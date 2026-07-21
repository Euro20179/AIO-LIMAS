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
		ctx.PP.Get("eventId", int64(0)).(int64),
		ctx.PP.Get("timezone", us.DefaultTimeZone).(string),
		ctx.PP["price"].(float64),
		ctx.PP["currency"].(string),
	)

	ctx.W.WriteHeader(200)
}

func DeleteTransaction(ctx RequestContext) {
	if err := db.DeleteTransaction(ctx.Uid, ctx.PP.Get("id", 0).(int64)); err != nil {
		util.WError(ctx.W, 500, "Failed to delete transaction: %s\n", err.Error())
	} else {
		success(ctx.W)
	}
}

func EditTransaction(ctx RequestContext) {
	t, err := db.GetTransaction(ctx.Uid, ctx.PP.Get("id", 0).(int64))
	if err != nil {
		util.WError(ctx.W, 500, "Unable to get transaction: %s\n", err.Error())
		return
	}

	if price := ctx.PP.Get("price", nil); price != nil {
		t.Price = price.(float64)
	}

	if currency := ctx.PP.Get("currency", nil); currency != nil {
		t.Currency = currency.(string)
	}

	if eventId := ctx.PP.Get("eventId", nil); eventId != nil {
		t.EventId = eventId.(int64)
	}

	if eventId := ctx.PP.Get("itemId", nil); eventId != nil {
		t.ItemId = eventId.(int64)
	}

	db.UpdateTransaction(ctx.Uid, &t)
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
