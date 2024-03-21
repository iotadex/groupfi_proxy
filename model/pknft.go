package model

func GetIssuerByNftid(nftid string) (string, string, error) {
	row := db.QueryRow("select `address`,`pk` from `issuer` where `nftid`=?", nftid)
	var addr, pk string
	if err := row.Scan(&addr, &pk); err != nil {
		return "", "", err
	}
	return addr, pk, nil
}
