#!/usr/bin/env bash

set -ex

source stackdriver-tools/ci/tasks/utils.sh

# BOSH and CF config
check_param bosh_director_address
check_param bosh_user
check_param bosh_password
check_param bosh_ca_cert

# CF settings
check_param cf_api_url
check_param firehose_username
check_param firehose_password

# Google network settings
check_param google_zone
check_param google_region
check_param network
check_param private_subnetwork

# Google service account settings
check_param cf_service_account_json
check_param ssh_user
check_param ssh_key

semver=`cat version-semver/number`

echo "Configuring SSH"
echo -e "${ssh_key}" > /tmp/${ssh_user}.key
chmod 700 /tmp/${ssh_user}.key

echo "Connecting to SSH bastion..."
ssh -4 -D 5000 -fNC bosh@${ssh_bastion_address} -i /tmp/${ssh_user}.key -o StrictHostKeyChecking=no
export BOSH_ALL_PROXY=socks5://localhost:5000

echo "Using BOSH CLI version..."
bosh2 --version
export BOSH_CLIENT=${bosh_user}
export BOSH_CLIENT_SECRET=${bosh_password}
export BOSH_ENVIRONMENT=https://${bosh_director_address}:25555
export BOSH_CA_CERT=${bosh_ca_cert}

echo "Targeting BOSH director..."
bosh2 login -n
bosh2 env

echo "Uploading nozzle release..."
bosh2 upload-release stackdriver-tools-artifacts/*.tgz

nozzle_manifest_name=stackdriver-nozzle.yml
cat > ${nozzle_manifest_name} <<EOF
---

name: stackdriver-nozzle-ci

releases:
- name: stackdriver-tools
  version: ${semver}

jobs:
- name: stackdriver-nozzle
  instances: 1
  networks:
    - name: private
  resource_pool: common
  templates:
    - name: stackdriver-nozzle
      release: stackdriver-tools
    - name: google-fluentd
      release: stackdriver-tools
    - name: stackdriver-agent
      release: stackdriver-tools
  properties:
    firehose:
      endpoint: ${cf_api_url}
      events_to_stackdriver_logging: LogMessage,Error,HttpStartStop,CounterEvent,ValueMetric,ContainerMetric
      events_to_stackdriver_monitoring: CounterEvent,ValueMetric,ContainerMetric
      username: ${firehose_username}
      password: ${firehose_password}
      skip_ssl: true
      newline_token: ∴
    gcp:
      project_id: ${cf_project_id}
    credentials:
      application_default_credentials: '${cf_service_account_json}'
    nozzle:
      debug: true

compilation:
  workers: 6
  network: private
  reuse_compilation_vms: true
  cloud_properties:
    zone: ${google_zone}
    machine_type: n1-standard-8
    root_disk_size_gb: 100
    root_disk_type: pd-ssd
    preemptible: true

resource_pools:
  - name: common
    network: private
    stemcell:
      name: bosh-google-kvm-ubuntu-trusty-go_agent
      version: latest
    cloud_properties:
      zone: ${google_zone}
      machine_type: n1-standard-4
      root_disk_size_gb: 20
      root_disk_type: pd-standard
  - name: nozzle
    network: private
    stemcell:
      name: bosh-google-kvm-ubuntu-trusty-go_agent
      version: latest
    cloud_properties:
      zone: ${google_zone}
      machine_type: n1-standard-4
      root_disk_size_gb: 20
      root_disk_type: pd-standard

networks:
  - name: private
    type: manual
    subnets:
    - range: 192.168.0.0/16
      reserved:
      - 192.168.0.0-192.168.1.15
      gateway: 192.168.0.1
      cloud_properties:
        zone: ${google_zone}
        network_name: ${network}
        subnetwork_name: ${private_subnetwork}
        ephemeral_external_ip: false
        tags:
          - stackdriver-nozzle-internal
          - internal
          - no-ip

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000

EOF

bosh2 -n deploy ${nozzle_manifest_name}

# Move release and its SHA256
mv stackdriver-tools-artifacts/*.tgz candidate/latest.tgz
mv stackdriver-tools-artifacts-sha256/*.tgz.sha256 candidate/latest.tgz.sha256
