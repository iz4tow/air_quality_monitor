#!/bin/bash

#install airmon
sudo mkdir -p /opt/airmon
sudo mv data_logger /opt/airmon

#Create airmon service
sudo tee /etc/systemd/system/airmon.service <<EOF
[Unit]
Description=Franco Air Quality Monitor Hub
After=network.target auditd.service

[Service]
WorkingDirectory=/opt/airmon/
ExecStart=/opt/airmon/data_logger
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable --now airmon

#Install Grafana
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get update
sudo apt-get install -y grafana
# Create provisioning directories
sudo rm -rf /etc/grafana/provisioning/
sudo mkdir -p /etc/grafana/provisioning/datasources
sudo mkdir -p /etc/grafana/provisioning/dashboards
sudo mkdir -p /var/lib/grafana/dashboards
# Creating provisioning files
sudo tee /etc/grafana/provisioning/datasources/datasource.yaml << EOF
apiVersion: 1
datasources:
  - name: SQLite
    type: frser-sqlite-datasource
    access: proxy
    jsonData:
      path: /opt/airmon/sensor_data.db
    basicAuth: false
    isDefault: true
    uid: "ae9yrqiwuhhc0f"
EOF
sudo tee /etc/grafana/provisioning/dashboards/dashboard.yaml << EOF
apiVersion: 1
providers:
  - name: Default
    orgId: 1
    folder: ""
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /var/lib/grafana/dashboards
EOF
sudo cp dashboard.json /var/lib/grafana/dashboards/
sudo cp dashboardrt.json /var/lib/grafana/dashboards/
# Enable Grafana on boot
sudo systemctl enable grafana-server
grafana-cli plugins install frser-sqlite-datasource
# Restart Grafana to apply provisioning
sudo systemctl restart grafana-server

while true; do
    read -p "Do you want to configure Whatsapp Sender? [Y/N] " answer
    case $answer in
        [Yy]) 
            read -p "Insert Whatsapp number with country code without + (es: italian number 3334455666 -> 393334455666): " whats_number
sudo tee /etc/systemd/system/airmon_alarm.service <<EOF
[Unit]
Description=Franco Air Quality Monitor Alarm
After=network.target auditd.service

[Service]
WorkingDirectory=/opt/airmon/
ExecStart=/opt/airmon/whatsapp_logger -number $whats_number
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
cd /opt/airmon/whatsapp
echo "Please scan QR code with WhatsApp on your phone to link this device"
whatsapp/whatsapp_login
sudo systemctl enable --now airmon_alarm
echo "System installed and configured successfully!"
beak



sudo systemctl enable airmon_alarm
            ;;
            

        [Nn])
            echo "You chose No."
echo "To enable alarm notification:"
echo "sudo su"
echo "cd /opt/airmon"
echo "whatsapp/whatsapp_login"
echo
echo "To install alarm service:"
echo "sudo tee /etc/systemd/system/airmon_alarm.service <<EOF"
echo "[Unit]"
echo "Description=Franco Air Quality Monitor Alarm"
echo "After=network.target auditd.service"
echo
echo "[Service]"
echo "WorkingDirectory=/opt/airmon/"
echo "ExecStart=/opt/airmon/whatsapp_logger -number <YOUR NUMBER WITH COUNTRYCODE WITHOUT +, es: 393334455666>"
echo "Restart=on-failure"
echo
echo "[Install]"
echo "WantedBy=multi-user.target"
echo "EOF"
echo "sudo systemctl enable --now airmon_alarm"
echo
echo "System installed and configured successfully!"
break
            ;;
        *)
            echo "Invalid input. Please enter Y or N."
            ;;
    esac

