#!/bin/sh

sqlboiler  --wipe -p sqlboilerPSQL -o pkg/api/adapters/dishRepo/sqlboilerPSQL psql || exit 1
go test ./pkg/api/adapters/dishRepo/sqlboilerPSQL