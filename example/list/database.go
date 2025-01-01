package main

type userData struct {
	name string
}

type userDatabase struct {
	users map[string]userData
}

func (db *userDatabase) Create(data userData) {
	db.users[data.name] = data
}

func (db *userDatabase) Read(name string) (userData, bool) {
	data, ok := db.users[name]
	return data, ok
}

func (db *userDatabase) Update(data userData) {
	db.users[data.name] = data
}

func (db *userDatabase) Delete(name string) {
	delete(db.users, name)
}
