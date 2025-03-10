# [<en>] general cluster parameters (ClusterConfiguration)
# [<en>] version of the Deckhouse API
# [<ru>] секция с общими параметрами кластера (ClusterConfiguration)
# [<ru>] используемая версия API Deckhouse Platform
apiVersion: deckhouse.io/v1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: ClusterConfiguration
# [<en>] type of the infrastructure: bare metal (Static) or Cloud (Cloud)
# [<ru>] тип инфраструктуры: bare metal (Static) или облако (Cloud)
clusterType: Cloud
# [<en>] cloud provider-related settings
# [<ru>] параметры облачного провайдера
cloud:
  # [<en>] type of the cloud provider
  # [<ru>] используемый облачный провайдер
  provider: AWS
  # [<en>] prefix to differentiate cluster objects (can be used, e.g., in routing)
  # [<ru>] префикс для объектов кластера для их отличия (используется, например, при маршрутизации)
  prefix: "cloud-demo"
# [<en>] address space of the cluster's Pods
# [<ru>] адресное пространство Pod’ов кластера
podSubnetCIDR: 10.111.0.0/16
# [<en>] address space of the cluster's services
# [<ru>] адресное пространство для service’ов кластера
serviceSubnetCIDR: 10.222.0.0/16
# [<en>] Kubernetes version to install
# [<ru>] устанавливаемая версия Kubernetes
kubernetesVersion: "1.21"
# [<en>] cluster domain (used for local routing)
# [<ru>] домен кластера
clusterDomain: "cluster.local"
---
# [<en>] section for bootstrapping the Deckhouse cluster (InitConfiguration)
# [<en>] version of the Deckhouse API
# [<ru>] секция первичной инициализации кластера Deckhouse (InitConfiguration)
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: InitConfiguration
# [<en>] Deckhouse parameters
# [<ru>] секция с параметрами Deckhouse
deckhouse:
  # [<en>] the release channel in use
  # [<ru>] используемый канал обновлений
  releaseChannel: Stable
  configOverrides:
    global:
      modules:
        # [<en>] template that will be used for system apps domains within the cluster
        # [<en>] e.g., Grafana for %s.example.com will be available as grafana.example.com
        # [<ru>] шаблон, который будет использоваться для составления адресов системных приложений в кластере
        # [<ru>] например, Grafana для %s.example.com будет доступна на домене grafana.example.com
        publicDomainTemplate: "%s.example.com"
---
# [<en>] section containing the parameters of the cloud provider
# [<en>] version of the Deckhouse API
# [<ru>] секция, описывающая параметры облачного провайдера
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: AWSClusterConfiguration
# [<en>] pre-defined layout from Deckhouse
# [<ru>] layout — архитектура расположения ресурсов в облаке
layout: WithNAT
# [<en>] AWS access parameters
# [<ru>] параметры доступа к облаку AWS
provider:
  providerAccessKeyId: *!CHANGE_MYACCESSKEY*
  providerSecretAccessKey: *!CHANGE_mYsEcReTkEy*
  # [<en>] cluster region
  # [<ru>] регион привязки кластера
  # [<en>] you might consider changing this
  # [<ru>] возможно, захотите изменить
  region: eu-central-1
withNAT:
  # [<en>] parameters of the bastion instance
  # [<ru>] параметры инстанса bastion-хоста
  bastionInstance:
    # [<ru>] зона в которой создавать bastion-хост
    # [<en>] bastion zone
    # [<en>] you might consider changing this
    # [<ru>] возможно, захотите изменить
    zone: eu-central-1a
    instanceClass:
      # [<en>] type of the bastion instance
      # [<ru>] тип используемого инстанса bastion-хоста
      # [<en>] you might consider changing this
      # [<ru>] возможно, захотите изменить
      instanceType: m5.large
      # [<en>] Amazon Machine Image id
      # [<ru>] id образа виртуальной машины в Amazon
      # [<en>] you might consider changing this
      # [<ru>] возможно, захотите изменить
      ami: ami-09a4a23815cdb5e06
# [<en>] parameters of the master node group
# [<ru>] параметры группы master-узлов
masterNodeGroup:
  # [<en>] number of replicas
  # [<en>] if more than 1 master node exists, control-plane will be automatically deployed on all master nodes
  # [<ru>] количество реплик мастера
  # [<ru>] если будет больше одного master-узла, то control-plane на всех master-узлах будет развернут автоматическии
  replicas: 1
  # [<en>] parameters of the VM image
  # [<ru>] параметры инстанса
  instanceClass:
    # [<en>] type of the instance
    # [<ru>] тип используемого инстанса
    # [<en>] you might consider changing this
    # [<ru>] возможно, захотите изменить
    instanceType: c5.large
    # [<en>] Amazon Machine Image id
    # [<ru>] id образа виртуальной машины в Amazon
    # [<en>] you might consider changing this
    # [<ru>] возможно, захотите изменить
    ami: ami-0fee04b212b7499e2
# [<en>] address space of the AWS cloud
# [<ru>] адресное пространство облака внутри AWS
vpcNetworkCIDR: "10.241.0.0/16"
# [<en>] address space of the cluster's nodes
# [<ru>] адресное пространство узлов кластера
nodeNetworkCIDR: "10.241.32.0/20"
# [<en>] public SSH key for accessing cloud nodes
# [<ru>] публичная часть SSH-ключа для доступа к узлам облака
sshPublicKey: ssh-rsa <SSH_PUBLIC_KEY>
