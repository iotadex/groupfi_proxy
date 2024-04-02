package model

func StoreBuyOrder(chain, txHash, user, edAddr, bech32Addr, fromAmount string, toAmount uint64) error {
	// 1. begin a transaction of mysql
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return err
	}

	// 2. insert into buy_order
	if _, err = tx.Exec("INSERT INTO `buy_order`(`chain`, `txhash`,`user`,`ed_addr`,`bech_addr`,`amount`) VALUES(?,?,?,?,?,?)", chain, txHash, user, edAddr, bech32Addr, fromAmount); err != nil {
		tx.Rollback()
		return err
	}

	// 3. insert pending send
	if err := InsertPendingSendSmrOrder(tx, bech32Addr, toAmount, SEND_BUY); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
