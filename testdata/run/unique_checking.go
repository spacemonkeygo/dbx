package main

import (
	"context"
)

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

func assert(x bool) {
	if !x {
		panic("assertion failed")
	}
}

var ctx = context.Background()

func main() {
	db, err := Open("sqlite3", ":memory:")
	erre(err)
	defer db.Close()

	_, err = db.Exec(db.Schema())
	erre(err)

	a, err := db.Create_A(ctx)
	erre(err)

	b1, err := db.Create_B(ctx, B_AId(a.Id))
	erre(err)
	c1, err := db.Create_C(ctx, C_Lat(0.0), C_Lon(0.0), C_BId(b1.Id))
	erre(err)

	b2, err := db.Create_B(ctx, B_AId(a.Id))
	erre(err)
	c2, err := db.Create_C(ctx, C_Lat(1.0), C_Lon(1.0), C_BId(b2.Id))
	erre(err)

	rows, err := db.All_A_B_C_By_A_Id_And_C_Lat_Less_And_C_Lat_Greater_And_C_Lon_Less_And_C_Lon_Greater(ctx,
		A_Id(a.Id),
		C_Lat(10.0), C_Lat(-10.0),
		C_Lon(10.0), C_Lon(-10.0))
	erre(err)

	assert(len(rows) == 2)

	assert(rows[0].A.Id == a.Id)
	assert(rows[0].B.Id == b1.Id)
	assert(rows[0].C.Id == c1.Id)
	assert(rows[0].C.Lat == 0)
	assert(rows[0].C.Lon == 0)

	assert(rows[1].A.Id == a.Id)
	assert(rows[1].B.Id == b2.Id)
	assert(rows[1].C.Id == c2.Id)
	assert(rows[1].C.Lat == 1)
	assert(rows[1].C.Lon == 1)
}
