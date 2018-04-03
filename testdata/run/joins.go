package main

import "context"

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

var ctx = context.Background()

func main() {
	db, err := Open("sqlite3", ":memory:")
	erre(err)
	defer db.Close()

	_, err = db.Exec(db.Schema())
	erre(err)

	user, err := db.Create_User(ctx)
	erre(err)

	aa, err := db.Create_AssociatedAccount(ctx,
		AssociatedAccount_UserPk(user.Pk))
	erre(err)

	sess, err := db.Create_Session(ctx,
		Session_UserPk(user.Pk))
	erre(err)

	rows, err := db.All_Session_Id_By_AssociatedAccount_Pk(ctx,
		AssociatedAccount_Pk(aa.Pk))
	erre(err)

	if len(rows) != 1 || rows[0].Id != sess.Id {
		panic("invalid")
	}
}
