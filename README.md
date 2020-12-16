**OddEye Local Proxy**
--------------

OddEye Proxy is a small and high performance proxy application that act as moc OddEye API server for agents running in internet restricted environments.  
It accepts requests from oe-agent, stores metrics in memory or RabbitMQ and push it to real OddEye API servers. 

#### **Installation**
OddEye local proxy is a simple Go application, which is ditributed as source code or pre-compilled, single file binary. 
Binaries are hosted at releases section, please download and sun the latest version for Linux , it does not have any external dependencies

#### **Build from soure**
To build OddEye proxy from source, you need Go sdk and binaries . Building from source fo Linux is  easy with only few simple steps. 

```bash
git clone https://github.com/oddeyeco/oe-go-proxy.git
cd oe-go-proxy 
export GOPATH=`pwd`
go get github.com/streadway/amqp
go get gopkg.in/ini.v1
go build -o oe-go-proxy src/start.go 
mkdir /usr/local/oe-go-proxy

mv oe-go-proxy /usr/local/oe-go-proxy/
mv src/config.ini /usr/local/oe-go-proxy/
cd /usr/local/oe-go-proxy/
chmod 755 oe-go-proxy
./oe-go-proxy
```

We recommend executing oe-go-proxy via system.d. Below are few simple steps to make systemd service and run it
OddEye proxy does not need any special privileges so it is safer to run it a regular system user. 
```ini
useradd -r -s /bin/false oddeye
```
Create `/etc/systemd/system/oe-go-proxy.service` file with following content


```ini
[Unit]
Description=OddEye-Proxy
Documentation=https://www.oddeye.co/documentation/
Wants=network-online.target
After=network-online.target

[Service]
User = oddeye
Group = oddeye
WorkingDirectory=/usr/local/oe-go-proxy
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/local/sbin/oe-go-proxy
KillMode=process
KillSignal=SIGINT
LimitNOFILE=infinity
LimitNPROC=infinity
Restart=on-failure
RestartSec=2
StartLimitBurst=3
StartLimitIntervalSec=10
TasksMax=infinity

[Install]
WantedBy=multi-user.target
```

Reload systemd and start service. 

```bash
systemctl daemon-reload 
systemctl enable oe-go-proxy.service
systemctl restart oe-go-proxy.service
```

#### **Configuration**

OddEye proxy uses simple and intuitive config file which ships with a source. Example config is below. 
config file should be in same folder as oe-go-proxy binary. Just set up few parameters and run the service. 

```ini
[main]
listen : 127.0.0.1:8080
remote : https://api.oddeye.co/oddeye-barlus/put/tsdb
dispatchers : 20
queuesize : 5000000
internalqueue : false

[server]
serverauth : true
serveruser : admin
serverpass : $ecretPaSs

[client]
clientauth : true
clientuser : admin
clientpass : $ecretPaSs

[rabbit]
rabbiturl : amqp://admin:admin@localhost:5672/
rabbitqueue : oemetrics

[monitoring]
enabled : true
url:  http://127.0.0.1:9191
user: admin
pass: $ecretPaSs
```

OddEye proxy support caching of messages in memory and RabbitMQ server. 
If due to any reason proxy will lose connection to main OddEye servers it will cache `queuesize` 
amount of messages in memory and if you use RabbitMQ you can cache as many messages and you have disk space for RabbitMQ. 

If you prefer not to use RabbitMQ, then OddEye proxy will cache `queuesize` amount of messages. Local queue is FIFO aof if there are more than `queuesize`, 
old one will be replaced with new messages.   

#### **Install RabbitMQ**

Now we need to install RabbitMQ. If you already have RabbitMQ installation at your environment, you can skip this steps. 

**Debian / Ubuntu** 

```bash
apt-get install rabbitmq
```

**CentOS / Redhat** 
```bash
yum -y install epel-release
yum -y update
yum -y install erlang socat
wget https://www.rabbitmq.com/releases/rabbitmq-server/${VERSION}/rabbitmq-server-${VERSION}.noarch.rpm
rpm --import https://www.rabbitmq.com/rabbitmq-release-signing-key.asc
rpm -Uvh rabbitmq-server-${VERSION}.noarch.rpm
systemctl start rabbitmq-server
systemctl enable rabbitmq-server
```

**Enable RabbitMQ web console**

```bash
rabbitmq-plugins enable rabbitmq_management
```

**Create admin user** 

```bash
rabbitmqctl add_user admin StrongPassword
rabbitmqctl set_user_tags admin administrator
rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"
```
You can manage RabbitMQ here : `http://Your_Server_IP:15672/`
