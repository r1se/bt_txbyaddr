package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func NewDB(host, port, user, pass, dbname, ssl string) *sql.DB {

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s",
		host,
		port,
		user,
		pass,
			ssl,
	))
	if err != nil {
		panic(err)
	}

	_, _ = db.Exec("CREATE DATABASE " + dbname)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`address(
  			"addr" varchar(255) UNIQUE,
			"latestblock"  varchar(255),
			"time" timestamp)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
		`transact(
			"addr" varchar(255) REFERENCES address(addr) on delete cascade on update cascade,
  			"txhash" varchar(255) NOT NULL,			
			"raw" text DEFAULT NULL,
			"block" text DEFAULT NULL,
  			"blockhash" text DEFAULT NULL,
			"blockheight" varchar(255) DEFAULT NULL,
  			"blocktime" timestamp DEFAULT NULL)`)
	if err != nil {
		panic(err)
	}

	return db
}

func InsertAddress(db *sql.DB, txhash string, blockhash string, inout ...interface{}) error {

	stmt, err := db.Prepare("INSERT INTO address VALUES($1,$2, to_timestamp($3));")
	if err != nil {
		log.Println("stmt err %v \n", err)
		return err
	}
	defer stmt.Close()

	for _, interf:= range inout {
		if t, ok := interf.([]*Inputs); ok {
			for _, input := range t {
				if input.PrevOut != nil {
					if input.PrevOut.Addr != "" {
						stmt.Exec(input.PrevOut.Addr, blockhash, myTime)
						if err != nil {
							log.Println("stmt err %v \n", err)
							return err
						}
						mapmutex.Lock()
						accounts[txhash] = append(accounts[txhash], input.PrevOut.Addr)
						mapmutex.Unlock()
					}
				}
			}
		}

		if t, ok := interf.([]*Out); ok {
			for _, out := range t {
				if out.Addr != "" {
					stmt.Exec(out.Addr , blockhash, myTime)
					if err != nil {
						log.Println("stmt err %v \n", err)
						return err
					}
					mapmutex.Lock()
					accounts[txhash] = append(accounts[txhash], out.Addr)
					mapmutex.Unlock()
				}
			}
		}
	}

	return nil
}

func InsertTransactions(db *sql.DB, txs []toDB) error {
	log.Println("Start insert transaction.")
	stmt, err := db.Prepare(
		"	INSERT INTO transact(addr, txhash, raw, block, blockhash, blockheight, blocktime) VALUES($1, $2, $3, $4, $5, $6, to_timestamp($7))")
	if err != nil {
		log.Println("stmt err %v \n", err)
	}
	defer stmt.Close()

	for _, tz := range txs {
		for _,addr:= range accounts[tz.Tx.Hash]{
			res, err := stmt.Exec(addr,
				tz.Tx.Hash,
				fmt.Sprintf("%v", *tz.Tx),
				fmt.Sprintf("%v", *tz.Block),
				tz.Block.Hash, tz.Block.Height, tz.Block.Time)
			if err != nil {
				log.Println("stmt err %v \n", res, err)
			}
		}
		mapmutex.Lock()
		delete(accounts, tz.Tx.Hash)
		mapmutex.Unlock()
	}
	log.Println("Transaction inserted.")
	return err

}
