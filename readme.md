go get .
go run .

psql -U postgres -d squirm -f back_end/database/create_users.sql
psql -U postgres -d squirm -f back_end/database/create_chatrooms.sql
psql -U postgres -d squirm -f back_end/database/create_chatroom_members.sql

psql -U postgres -d squirm_test -f back_end/database/create_users.sql
psql -U postgres -d squirm_test -f back_end/database/create_chatrooms.sql
psql -U postgres -d squirm_test -f back_end/database/create_chatroom_members.sql
