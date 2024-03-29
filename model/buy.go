package model

func StoreBuyOrder(chain, txHash, user, pubkey, addr, amount string) error {
	_, err := db.Exec("INSERT INTO `groupfi`.`buy_order`(`chain`, `txhash`,`user`,`pubkey`,`bech_addr`,`amount`) VALUES(?,?,?,?,?,?)", chain, txHash, user, pubkey, addr, amount)
	return err
}
