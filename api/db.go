package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func NewDB(host, port, user, pass, dbname string) *sql.DB {

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s ",
		host,
		port,
		user,
		pass,
	))
	if err != nil {
		panic(err)
	}

	_, _ = db.Exec("CREATE DATABASE " + dbname)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`address(
  			"addr" varchar(255) UNIQUE)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`transact(
  			"txhash" varchar(255) NOT NULL UNIQUE,
			"addr" varchar(255) REFERENCES address(addr),
			"raw" text DEFAULT NULL,
			"block" text DEFAULT NULL,
  			"blockhash" text DEFAULT NULL,
			"blockheight" varchar(255) DEFAULT NULL,
  			"blocktime" bigint DEFAULT NULL)`)
	if err != nil {
		panic(err)
	}




	return db
}

func InsertAddress(db *sql.DB, txhash string, inout interface{}) error {

	stmt, err := db.Prepare("INSERT INTO address VALUES($1);")
	if err != nil {
		log.Println("stmt err %v \n", err)
	}

	if t, ok := inout.([]*Inputs); ok {
		for _, input := range t {
			if input.PrevOut != nil {
				if input.PrevOut.Addr != "" {
					mapmutex.Lock()
					accounts[txhash] = append(accounts[txhash], input.PrevOut.Addr)
					mapmutex.Unlock()
					stmt.Exec(input.PrevOut.Addr)
					if err != nil {
						log.Println("stmt err %v \n", err)
						return err
					}

				}
			}
		}
	}

	if t, ok := inout.([]*Out); ok {
		for _, out := range t {
			if out.Addr != "" {
				mapmutex.Lock()
				accounts[txhash] = append(accounts[txhash], out.Addr)
				mapmutex.Unlock()
				stmt.Exec(out.Addr)
				if err != nil {
					log.Println("stmt err %v \n", err)
					return err
				}

			}
		}
	}

	return nil
}

func InsertTransactions(db *sql.DB, txs []toDB) error {
	stmt, err := db.Prepare(
		"	INSERT INTO transact(txhash, addr,raw, block, blockhash, blockheight, blocktime) VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (txhash) DO UPDATE SET  addr = EXCLUDED.addr,raw = EXCLUDED.raw, block = EXCLUDED.block, blockhash = EXCLUDED.blockhash, blockheight = EXCLUDED.blockheight, blocktime = EXCLUDED.blocktime")
	if err != nil {
		log.Println("stmt err %v \n", err)
	}
	defer stmt.Close()

	for _, tz := range txs {

		for k, v := range accounts {
			for _, element := range v {
				res, err := stmt.Exec(k, element, fmt.Sprintf("%v", *tz.Tx), fmt.Sprintf("%v", *tz.Block), tz.Block.Hash, tz.Block.Height, tz.Block.Time)
				if err != nil {
					log.Println("stmt err %v \n", res, err)
				}
			}
		}


	}
	return err

}
