#!/bin/bash
rm -rf ../RPI-AQI-Hub
rm -f ../RPI-AQI-Hub.tar.gz
mkdir -p ../RPI-AQI-Hub/whatsapp
sudo apt-get install gcc-arm* -y
sudo apt install jq
export GOOS=linux; \
export GOARCH=arm; \
export GOARM=7; \
export CC=arm-linux-gnueabi-gcc; \
CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" --trimpath  data_logger.go
CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" --trimpath  whatsapp_logger.go
cd whatsapp
CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" --trimpath send_whatsapp.go
CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" --trimpath whatsapp_login.go
cd ..

curl -X GET "http://localhost:3000/api/dashboards/uid/be9yswpcby39ca" -H "Authorization: Bearer glsa_0a7GSB9AI7Dwuuc6cbQP0P0fszRrVsqO_76a78700"
curl -X GET "http://localhost:3000/api/dashboards/uid/be9yswpcby39ca" -H "Authorization: Bearer glsa_0a7GSB9AI7Dwuuc6cbQP0P0fszRrVsqO_76a78700" > dashboard_toclean.json
jq '.dashboard' "dashboard_toclean.json" > "dashboard.json"
curl -X GET "http://localhost:3000/api/dashboards/uid/fea5brmbiqyo0c" -H "Authorization: Bearer glsa_0a7GSB9AI7Dwuuc6cbQP0P0fszRrVsqO_76a78700"
curl -X GET "http://localhost:3000/api/dashboards/uid/fea5brmbiqyo0c" -H "Authorization: Bearer glsa_0a7GSB9AI7Dwuuc6cbQP0P0fszRrVsqO_76a78700" > dashboardrt_toclean.json
jq '.dashboard' "dashboardrt_toclean.json" > "dashboardrt.json"

rm -f dashboard_toclean.json
mv data_logger ../RPI-AQI-Hub
mv whatsapp_logger ../RPI-AQI-Hub
mv whatsapp/send_whatsapp ../RPI-AQI-Hub/whatsapp
mv whatsapp/whatsapp_login.go ../RPI-AQI-Hub/whatsapp
cp dashboard.json ../RPI-AQI-Hub
cp dashboardrt.json ../RPI-AQI-Hub
cp rpi_hub_install.sh ../RPI-AQI-Hub

cd ../RPI-AQI-Hub
tar -czvf ../RPI-AQI-Hub.tar.gz *
