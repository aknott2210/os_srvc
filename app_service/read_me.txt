Example Consul start:
sudo ./app_service -service install -app /etc/clarify-tools/consul/consul -serviceName consul -config /etc/consul.d/ -configFlag -config-dir

Example Nomad start:
sudo ./app_service -service install -app /etc/clarify-tools/nomad/nomad -serviceName nomadtest -config /etc/nomad.d/ -configFlag -config