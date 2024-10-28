#!/bin/bash

go build -o bookings cmd/web/*.go && ./bookings
./bookings -dbname=hotel-bookings -dbuser=user -cache=false production=false