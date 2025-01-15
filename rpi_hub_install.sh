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
# Enable Grafana on boot
sudo systemctl enable grafana-server
grafana-cli plugins install frser-sqlite-datasource
# Restart Grafana to apply provisioning
sudo systemctl restart grafana-server
echo "System installed and configured successfully!"
