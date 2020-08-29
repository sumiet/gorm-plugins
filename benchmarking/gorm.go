package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

var gm *gorm.DB

func registerGORM() {
	st := NewSuite("gorm")
	st.InitF = func() {
		st.AddBenchmark("Insert", 2000*ORM_MULTI, GormInsert)
		st.AddBenchmark("MultiInsert 100 row", 500*ORM_MULTI, GormInsertMulti)
		st.AddBenchmark("Update", 2000*ORM_MULTI, GormUpdate)
		st.AddBenchmark("Read", 4000*ORM_MULTI, GormRead)
		st.AddBenchmark("MultiRead limit 100", 2000*ORM_MULTI, GormReadSlice)

		gm, _ = gorm.Open("mysql", ORM_SOURCE)
		// if err != nil {
		// 	panic(err)
		// }
		gm.DB().SetMaxIdleConns(ORM_MAX_IDLE)
		gm.DB().SetMaxOpenConns(ORM_MAX_CONN)

		gm.SingularTable(true)
		gm.AutoMigrate(&Model{})
	}
}

func GormInsert(b *B) {
	var m *Model
	wrapExecute(b, func() {
		initDB()

		m = NewModel()
	})

	for i := 0; i < b.N; i++ {
		m.Id = 0
		gm.Create(m)
	}
}

func GormInsertMulti(b *B) {
	// var ms []Model
	// wrapExecute(b, func() {
	// 	initDB()
	//
	// 	ms = make([]Model, 0, 100)
	// 	for i := 0; i < 100; i++ {
	// 		ms = append(ms, *NewModel())
	// 	}
	// })
	//
	// for i := 0; i < b.N; i++ {
	// 	gm.Create(&ms)
	// }
	panic(fmt.Errorf("Not support multi insert"))
}

func GormUpdate(b *B) {
	var m *Model
	wrapExecute(b, func() {
		initDB()
		m = NewModel()
		gm.Create(m)
	})

	for i := 0; i < b.N; i++ {
		gm.Save(m)
	}
}

func GormRead(b *B) {
	var m *Model
	wrapExecute(b, func() {
		initDB()
		m = NewModel()
		gm.Create(m)
	})

	for i := 0; i < b.N; i++ {
		gm.First(m, m.Id)
	}
}

func GormReadSlice(b *B) {
	var m *Model
	wrapExecute(b, func() {
		initDB()
		m = NewModel()
		for i := 0; i < 100; i++ {
			m.Id = 0
			gm.Create(m)
		}
	})
	for i := 0; i < b.N; i++ {
		var models []*Model
		gm.Where("id > ?", 0).Limit(100).Find(&models)
	}
}
