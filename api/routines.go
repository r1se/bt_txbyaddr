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

	i, err := strconv.Atoi(Config.Checklastblocktimeout)
	if err != nil {
		panic(err)
	}
	timer := time.NewTimer(time.Second * time.Duration(i))
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
			if lasthashblock != latestblock.Hash {
				blockIdChan <- latestblock.Hash
			}
		}
	}
}

func BlockDetail(blockIdChan <-chan string, TsChan chan<- []toDB) {

	for {
		select {
		case blockIds := <-blockIdChan:

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
				return
			}

			sliceTrans := []toDB{}
			for _, tx := range infoBlock.Tx {
				tmp := infoBlock
				tmp.Tx = nil
				tmp.TxIndexes = nil

				InsertAddress(db, tx.Hash, tx.Inputs)
				InsertAddress(db, tx.Hash, tx.Out)

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

	switch ok, err := ValidA58(data); {
	case ok:
	case err == nil:
		http.Error(w, "Invalid address ", http.StatusBadRequest)
		return
	default:
		http.Error(w, "Invalid address "+err.Error(), http.StatusBadRequest)
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
