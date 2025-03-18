#!/bin/bash
shellExit()
{
if [ $1 -eq 1 ]; then
    printf "\nfailed!!!\n\n"
    exit 1
fi
}


printf "\nRegenerating file\n\nstart build markdown and model file"


# docker run --name mariadb -p 3306:3306 -e MYSQL_ROOT_PASSWORD=xxxx -v /data/mariadb/data:/var/lib/mysql -d mariadb:10.1.21
#go run -v ./cmd/db_gen/main.go -type sqlite -dsn storage/database/db.db -name main -table pre -prefix pre_ -savedir test
#go run -v ./cmd/db_gen/main.go -type mysql -dsn "root:root@tcp(192.168.138.190:3306)/main?charset=utf8mb4&parseTime=true&loc=Local" -name main -table pre -prefix pre_ -savedir test


#  scripts/gormgen.sh sqlite storage/database/db.db  main  pre_  pre_  test

time go run -v ./cmd/db_gen/main.go  -type "$1" -dsn "$2" -name $3 -table "$4" -prefix "$5" -savedir "$6"

printf "\nBuild markdown and model file succeed!\n"

shellExit $?



printf "\ncreate curd code : \n"
time go build -o gormgen ./cmd/gorm_gen/main.go
shellExit $?

mv gormgen $GOPATH/bin/
shellExit $?

go generate ./...
shellExit $?

printf "\nFormatting code\n\n"
time go run -v ./cmd/mfmt/main.go
shellExit $?

printf "\nDone.\n\n"
