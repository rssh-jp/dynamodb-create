package main

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

func createTable(db *dynamo.DB, tableName string, info interface{}, isDelete bool) error {
	const (
		statusInitialize = iota + 1
		statusDelete
		statusCreate
		statusFinalize
	)

	status := statusInitialize

	for loop := true; loop; {
		switch status {
		case statusInitialize:
			log.Println("statusInitialize")
			tables, err := db.ListTables().All()
			if err != nil {
				return err
			}

			var isExist bool
			for _, table := range tables {
				if table == tableName {
					isExist = true
					break
				}
			}

			if isExist {
				if isDelete {
					status = statusDelete
				} else {
					status = statusFinalize
				}
			} else {
				status = statusCreate
			}

		case statusDelete:
			log.Println("statusDelete")
			err := db.Table(tableName).DeleteTable().Run()
			if err != nil {
				return err
			}

			err = waitDeleteTable(db, tableName)
			if err != nil {
				return err
			}

			status = statusCreate

		case statusCreate:
			log.Println("statusCreate")
			err := db.CreateTable(tableName, info).Run()
			if err != nil {
				return err
			}

			err = waitCreateTable(db, tableName)
			if err != nil {
				return err
			}

			status = statusFinalize

		case statusFinalize:
			log.Println("statusFinalize")
			loop = false

		default:
			loop = false
		}
	}

	return nil
}

func waitCreateTable(db *dynamo.DB, tableName string) error {
	return waitExecute(func() (bool, error) {
		desc, err := db.Table(tableName).Describe().Run()
		if err != nil {
			return false, err
		}

		if desc.Status == dynamo.ActiveStatus {
			return true, nil
		}

		return false, nil
	})
}

func waitDeleteTable(db *dynamo.DB, tableName string) error {
	return waitExecute(func() (bool, error) {
		tables, err := db.ListTables().All()
		if err != nil {
			return false, err
		}

		var isExist bool

		for _, table := range tables {
			if table == tableName {
				isExist = true
				break
			}
		}

		if !isExist {
			return true, nil
		}

		return false, nil
	})
}

func waitExecute(funcIsEnd func() (bool, error)) error {
	timer := time.NewTimer(time.Second * 5)

loop:
	for {
		select {
		case <-timer.C:
			break loop
		}

		isEnd, err := funcIsEnd()
		if err != nil {
			return err
		}

		if isEnd {
			break loop
		}

		time.Sleep(time.Millisecond)
	}

	return nil
}

func main() {
	log.Println("START")
	defer log.Println("END")

	cfg := aws.NewConfig()
	cfg.WithEndpoint("http://dynamo:8000")
	//cfg.WithRegion("us-west-2")
	cfg.WithRegion("ap-northeast-1")
	cfg.WithCredentials(credentials.NewStaticCredentials("dummy", "dummy", "dummy"))

	db := dynamo.New(session.New(), cfg)

	tables, err := db.ListTables().All()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tables)

	type UserAction struct {
		UserID string `dynamo:"user_id,hash"`
	}

	err = createTable(db, "test", UserAction{}, true)
	if err != nil {
		log.Fatal(err)
	}
}
