package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetEvents(blockIdChan chan<- string) {

	var tmp string
	rtrow := db.QueryRow("SELECT DISTINCT latestblock, time FROM address ORDER BY time desc;")
	rtrow.Scan(&lasthashblock, &tmp)

	i, err := strconv.Atoi(Config.Checklastblocktimeout)
	if err != nil {
		panic(err)
	}
	myTimer = i
	timer := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timer.C:

			req, err := http.NewRequest("GET", "https://blockchain.info/ru/latestblock", nil)
			if err != nil {
				log.Printf("Чтение request " + err.Error())
				return
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("Ошибка при обращении к REST " + err.Error())
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				bodyBytes, _ := ioutil.ReadAll(resp.Body)
				log.Println("Ошибка работы с базой: " + string(bodyBytes))
			}

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Чтение тела " + err.Error())
				return
			}

			latestblock := struct {
				Hash       string `json:"hash"`
				Time       int    `json:"time"`
				BlockIndex int    `json:"block_index"`
				Height     int    `json:"height"`
				TxIndexes  []int  `json:"txIndexes"`
			}{}

			err = json.Unmarshal(bodyBytes, &latestblock)
			if err != nil {
				log.Printf("Unmarshal request blockchain error " + err.Error())
				return
			}
			myTime = latestblock.Time
			if lasthashblock != latestblock.Hash {
				lasthashblock = latestblock.Hash
				blockIdChan <- latestblock.Hash
			}
			timer.Reset(time.Second * time.Duration(i))
		}
	}
}

func BlockDetail(blockIdChan <-chan string, TsChan chan<- []toDB) {

	for {
		select {
		case blockIds := <-blockIdChan:
REQ:
			req, err := http.NewRequest("GET", "https://blockchain.info/ru/rawblock/"+blockIds, nil)
			if err != nil {
				log.Printf("Чтение request " + err.Error())
				return
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("Ошибка при обращении к REST " + err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				bodyBytes, _ := ioutil.ReadAll(resp.Body)
				log.Println("Ошибка работы с базой: " + string(bodyBytes))
			}

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("error body read " + err.Error())
				return
			}

			infoBlock := Block{}
			err = json.Unmarshal(bodyBytes, &infoBlock)
			if err != nil {
				log.Printf("error unmarshall " + err.Error())
				time.Sleep(time.Second * 6)
				goto REQ
			}

			sliceTrans := []toDB{}
			for _, tx := range infoBlock.Tx {

				InsertAddress(db, tx.Hash, blockIds, tx.Inputs, tx.Out)

				tmp := infoBlock
				tmp.Tx = nil
				tmp.TxIndexes = nil
				sliceTrans = append(sliceTrans, toDB{&tmp, tx})
			}

			TsChan <- sliceTrans
		}
	}
}

func commHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant make slice from body "+err.Error(), http.StatusBadRequest)
		return
	}


	if ok, _ := ValidA58(data);!ok {
	http.Error(w, "Invalid address "+string(data), http.StatusBadRequest)
	return
	}
	//SELECT * FROM transact WHERE addr = '1NDyJtNTjmwk5xPNhjgAMu4HDHigtobu1s';
	stmt, err := db.Prepare(`SELECT * FROM transact WHERE addr=$1`)
	if err != nil {
		http.Error(w, "SQL statement prepare error. "+err.Error(), http.StatusBadRequest)
		return
	}
	defer stmt.Close()

	rtrow, err := db.Query("SELECT * FROM transact WHERE addr='" + string(data) + "';")
	if err != nil {
		http.Error(w, "query error. "+err.Error(), http.StatusBadRequest)
		return
	}

	tmpme := []answer{}

	for rtrow.Next() {
		tmp := answer{}
		rtrow.Scan(&tmp.Txhash, &tmp.Addr, &tmp.Raw, &tmp.Block, &tmp.Blockhash, &tmp.Blockheight, &tmp.Blocktime)
		tmpme = append(tmpme, tmp)
	}
	rtrow.Close()

	answer, err := json.Marshal(tmpme)
	if err != nil {
		http.Error(w, "query error. "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write(answer)
	return
}
