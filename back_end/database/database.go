package database

type Database struct {
}

func (database *Database) Connect() {
	/*
		Connect w/ RDS
		Configure connection pool
	*/
}

func (database *Database) HealthCheck() {
	/*
		Check database reachable
	*/
}

func (database *Database) Close() {
	/*
		Release db resources
	*/
}
