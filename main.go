package main

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

func createTable(db *dynamo.DB, tableName string, isDelete bool) error {
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

			status = statusCreate
		case statusCreate:
			log.Println("statusCreate")
			type UserAction struct {
				UserID string `dynamo:"user_id,hash"`
			}
			err := db.CreateTable(tableName, UserAction{}).Run()
			if err != nil {
				return err
			}

			status = statusFinalize
		case statusFinalize:
			log.Println("statusFinalize")
			loop = false
		}
	}

	return nil
}

func waitDeleteTable(db *dynamo.DB, tableName string) error {
	timer := time.NewTimer(time.Second * 5)

loop:
	for {
		select {
		case <-timer.C:
			break loop
		}

		time.Sleep(time.Second)
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

	err = createTable(db, "test", true)
	if err != nil {
		log.Fatal(err)
	}
}
