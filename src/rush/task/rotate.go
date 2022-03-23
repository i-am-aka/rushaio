package task

import (
	"github.com/pkg/errors"
	"net/url"
	"database/sql"
)

var ErrNoProxyAvailable = errors.New("No Proxy Available")
var ErrDatabaseUnavailable = errors.New("Database Unavailable")

func (t *CheckoutTask) RotateProxy() error {
	if t.GetTelemetryContext("rotate") == "0" {
		return nil
	}

	CloseH2Conns(t.client)
	t.client.CloseIdleConnections()
	if t.Db == nil {
		return ErrDatabaseUnavailable
	}
	if t.ProxyGroup == "" {
		CloseH2Conns(t.client)
		t.client.CloseIdleConnections()
		return nil
	}


	// TODO proxy groups
	// TODO improve random distribution
	// TODO use non relational db

	var rows *sql.Rows
	var err error
	tx, err := t.Db.Begin()
	if err != nil {
	  t.LogDebug("failed to begin transaction: %+v", err)
	  return err
	}



	if len(t.ProxyGroup) == 0 {
		t.ProxyGroup = "default"
	}
	// 	stmt, err := (`SELECT proxy.url, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state in ('Rushing', 'Starting')) cnt FROM proxy WHERE cnt = 0 ORDER BY cnt LIMIT 1;`)
	// 	rows, err = stmt.Exec()
	// } else {
	stmt, err := tx.Prepare(`SELECT * from (SELECT proxy.url, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state in ('Rushing', 'Starting')) cnt FROM proxy WHERE cnt = 0 AND pgroup LIKE ? LIMIT 128) ORDER BY RANDOM() LIMIT 1;`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err = stmt.Query(t.ProxyGroup)
	// }

	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var proxyUrl string
		var cnt int
		rows.Scan(&proxyUrl, &cnt)
		if purl, err := url.Parse(proxyUrl); err == nil {
			defer func() {
				if err == nil {
					t.StatusCh <- StatusRotatingProxy
					t.Proxy = purl
					t.LogDebug("Rotated to %+v", t.Proxy.Host)
					t.resetClient(false)
					CloseH2Conns(t.client)
					t.client.CloseIdleConnections()
					jar := t.client.Jar
					t.client, _ = t.newHttpClient()
					t.client.Jar = jar
				}
			}()


			stmt, err := tx.Prepare("update tasks2 set proxy=? where id = ?")
			if err != nil {
				tx.Rollback()
			  t.LogDebug("failed to prepare db state update: %+v", err)
			  return err
			}
			defer stmt.Close()
			_, err = stmt.Exec(proxyUrl, t.Id)
			if err != nil {
				tx.Rollback()
			  t.LogDebug("failed to update state: %+v", err)
			  return err
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
			t.Metrics.Set("proxy_conn_fail", 0)
			return nil
		}
	}
	tx.Rollback()
	return ErrNoProxyAvailable
}