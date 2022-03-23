package task


func (t *CheckoutTask) VaultGet(host string, ident string) *string {
	if t.Db == nil {
		return nil
	}
	rows, err := t.Db.Query(`SELECT secret FROM vault WHERE key = ?;`, host + "_" + ident)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var value string
	for rows.Next() {
		rows.Scan(&value)
		return &value
	}
	return nil
}

func (t *CheckoutTask) VaultPut(host string, ident string, secret string) {
	tx, err := t.Db.Begin()
	if err != nil {
	  t.LogDebug("failed to begin transaction: %+v", err)
	  return
	}

	stmt, err := tx.Prepare("update vault set secret=? where key=?")
	if t.VaultGet(host, ident) == nil {
		stmt, err = tx.Prepare("insert into vault (secret, key) values (?, ?);")
	}

	_, err = stmt.Exec(secret, host + "_" + ident)
	if err != nil {
	  t.LogDebug("failed to update state: %+v", err)
	  return
	}

	// TODO flag old prox; create proxy_bans table. expiration strategy?
	// stmt, err := tx.Prepare("update proxy set =? where id = ?")
	// if err != nil {
	//   t.LogDebug("failed to prepare db state update: %+v", err)
	//   return err
	// }
	// _, err = stmt.Exec(proxyUrl, t.Id)
	// if err != nil {
	//   t.LogDebug("failed to update state: %+v", err)
	//   return err
	// }

	tx.Commit()
	stmt.Close()
}