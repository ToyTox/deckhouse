# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: AzureInstanceClass
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] instance class name
  # [<ru>] имя инстанс класса
  name: worker
# [<en>] Parameters of the VM instance
# [<ru>] параметры инстанса
spec:
  # [<en>] type of the instance
  # [<ru>] тип используемого инстанса
  # [<en>] you might consider changing this
  # [<ru>] возможно, захотите изменить
  machineSize: Standard_F4
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: NodeGroup
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] node group name
  # [<ru>] имя группы узлов
  name: worker
spec:
  # [<en>] parameters for provisioning the cloud-based VMs
  # [<ru>] параметры заказа облачных виртуальных машин
  cloudInstances:
    # [<en>] the reference to the InstanceClass object
    # [<ru>] ссылка на объект InstanceClass
    classReference:
      # [<en>] the object type
      # [<ru>] тип объекта
      kind: AzureInstanceClass
      # [<en>] the name of the InstanceClass object
      # [<ru>] имя InstanceClass-объекта
      name: worker
    # [<en>] the maximum number of instances for the group in each zone
    # [<ru>] максимальное количество инстансов в зоне
    maxPerZone: 1
    # [<en>] the minimum number of instances for the group in each zone
    # [<ru>] минимальное количество инстансов в зоне
    minPerZone: 1
    # [<en>] list of availability zones to create instances in
    # [<ru>] переопределение перечня зон, в которых создаются инстансы
    zones: ["1"]
  # [<en>] the type of nodes this group provides
  # [<ru>] тип узлов, которые представляет эта группа
  nodeType: Cloud
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: AzureInstanceClass
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] instance class name
  # [<ru>] имя инстанс класса
  name: system
# [<en>] Parameters of the VM instance
# [<ru>] параметры инстанса
spec:
  # [<en>] type of the instance
  # [<ru>] тип используемого инстанса
  # [<en>] you might consider changing this
  # [<ru>] возможно, захотите изменить
  machineSize: Standard_F4
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: NodeGroup
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] node group name
  # [<ru>] имя группы узлов
  name: system
spec:
  # [<en>] parameters for provisioning the cloud-based VMs
  # [<ru>] параметры заказа облачных виртуальных машин
  cloudInstances:
    # [<en>] the reference to the InstanceClass object
    # [<ru>] ссылка на объект InstanceClass
    classReference:
      # [<en>] the object type
      # [<ru>] тип объекта
      kind: AzureInstanceClass
      # [<en>] the name of the InstanceClass object
      # [<ru>] имя InstanceClass-объекта
      name: system
    # [<en>] the maximum number of instances for the group in each zone
    # [<ru>] максимальное количество инстансов в зоне
    maxPerZone: 1
    # [<en>] the minimum number of instances for the group in each zone
    # [<ru>] минимальное количество инстансов в зоне
    minPerZone: 1
    # [<en>] list of availability zones to create instances in
    # [<ru>] переопределение перечня зон, в которых создаются инстансы
    zones: ["1", "2"]
  # [<en>] specification of some of the fields that will be maintained in all nodes of the group
  # [<ru>] настройки Node-объектов в Kubernetes, которые будут добавлены после регистрации узла
  nodeTemplate:
    # [<en>] similar to the standard metadata.labels field
    # [<ru>] аналогично стандартному полю metadata.labels
    labels:
      node-role.deckhouse.io/system: ""
    # [<en>] similar to the .spec.taints field of the Node object
    # [<en>] only effect, key, value fields are available
    # [<ru>] аналогично полю .spec.taints из объекта Node
    # [<ru>] доступны только поля effect, key, value
    taints:
    - effect: NoExecute
      key: dedicated.deckhouse.io
      value: system
  # [<en>] the type of nodes this group provides
  # [<ru>] тип узлов, которые представляет эта группа
  nodeType: Cloud
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: AzureInstanceClass
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] instance class name
  # [<ru>] имя инстанс класса
  name: frontend
# [<en>] Parameters of the VM instance
# [<ru>] параметры инстанса
spec:
  # [<en>] type of the instance
  # [<ru>] тип используемого инстанса
  # [<en>] you might consider changing this
  # [<ru>] возможно, захотите изменить
  machineSize: Standard_F4
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: NodeGroup
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] node group name
  # [<ru>] имя группы узлов
  name: frontend
spec:
  # [<en>] parameters for provisioning the cloud-based VMs
  # [<ru>] параметры заказа облачных виртуальных машин
  cloudInstances:
    # [<en>] the reference to the InstanceClass object
    # [<ru>] ссылка на объект InstanceClass
    classReference:
      # [<en>] the object type
      # [<ru>] тип объекта
      kind: AzureInstanceClass
      # [<en>] the name of the InstanceClass object
      # [<ru>] имя InstanceClass-объекта
      name: frontend
    # [<en>] the maximum number of instances for the group in each zone
    # [<ru>] максимальное количество инстансов в зоне
    maxPerZone: 1
    # [<en>] the minimum number of instances for the group in each zone
    # [<ru>] минимальное количество инстансов в зоне
    minPerZone: 1
    # [<en>] list of availability zones to create instances in
    # [<ru>] переопределение перечня зон, в которых создаются инстансы
    zones: ["1", "2"]
  # [<en>] specification of some of the fields that will be maintained in all nodes of the group
  # [<ru>] настройки Node-объектов в Kubernetes, которые будут добавлены после регистрации узла
  nodeTemplate:
    # [<en>] similar to the standard metadata.labels field
    # [<ru>] аналогично стандартному полю metadata.labels
    labels:
      node-role.deckhouse.io/frontend: ""
    # [<en>] similar to the .spec.taints field of the Node object
    # [<en>] only effect, key, value fields are available
    # [<ru>] аналогично полю .spec.taints из объекта Node
    # [<ru>] доступны только поля effect, key, value
    taints:
    - effect: NoExecute
      key: dedicated.deckhouse.io
      value: frontend
  # [<en>] the type of nodes this group provides
  # [<ru>] тип узлов, которые представляет эта группа
  nodeType: Cloud
---
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1alpha1
# [<en>] type of the configuration section
# [<ru>] тип секции конфигурации
kind: IngressNginxController
# [<en>] data that helps uniquely identify the object
# [<ru>] данные, которые помогают однозначно идентифицировать объект
metadata:
  # [<en>] object name
  # [<ru>] имя объекта
  name: main
spec:
  # [<en>] the name of the Ingress class to use with the Ingress nginx controller
  # [<ru>] имя Ingress-класса для обслуживания Ingress NGINX controller
  ingressClass: "nginx"
  # [<en>] the way traffic goes to cluster from the outer network
  # [<ru>] способ поступления трафика из внешнего мира
  inlet: "LoadBalancer"
  # [<en>] describes on which nodes the component will be located
  # [<ru>] описывает, на каких узлах будет находиться компонент
  nodeSelector:
    node-role.deckhouse.io/frontend: ""
---
apiVersion: deckhouse.io/v1
kind: ClusterAuthorizationRule
metadata:
  name: admin
spec:
  # [<ru>] список учётных записей Kubernetes RBAC
  # [<en>] Kubernetes RBAC accounts list
  subjects:
  - kind: User
    name: admin@example.com
  # [<ru>] предустановленный шаблон уровня доступа
  # [<en>] pre-defined access template
  accessLevel: SuperAdmin
  # [<ru>] разрешить пользователю делать kubectl port-forward
  # [<en>] allow user to do kubectl port-forward
  portForwarding: true
---
# [<en>] section containing the parameters of the static user
# [<ru>] секция, описывающая параметры статического пользователя
# [<en>] version of the Deckhouse API
# [<ru>] используемая версия API Deckhouse
apiVersion: deckhouse.io/v1
kind: User
metadata:
  name: admin
spec:
  # [<en>] user e-mail
  # [<ru>] e-mail пользователя
  email: admin@example.com
  # [<en>] this is a hash of the password <GENERATED_PASSWORD>, generated  now
  # [<en>] generate your own or use it at your own risk (for testing purposes)
  # [<en>] echo "<GENERATED_PASSWORD>" | htpasswd -BinC 10 "" | cut -d: -f2
  # [<ru>] это хэш пароля <GENERATED_PASSWORD>, сгенерированного сейчас
  # [<ru>] сгенерируйте свой или используйте этот, но только для тестирования
  # [<ru>] echo "<GENERATED_PASSWORD>" | htpasswd -BinC 10 "" | cut -d: -f2
  # [<en>] you might consider changing this
  # [<ru>] возможно, захотите изменить
  password: <GENERATED_PASSWORD_HASH>
