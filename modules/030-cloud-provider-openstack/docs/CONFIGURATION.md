---
title: "Сloud provider — OpenStack: настройки"
---

Модуль автоматически включается для всех облачных кластеров развёрнутых в OpenStack.

Количество и параметры процесса заказа машин в облаке настраиваются в custom resource [`NodeGroup`](/modules/040-node-manager/cr.html#nodegroup) модуля node-manager, в котором также указывается название используемого для этой группы узлов instance-класса (параметр `cloudInstances.classReference` NodeGroup).  Instance-класс для cloud-провайдера OpenStack — это custom resource [`OpenStackInstanceClass`](cr.html#openstackinstanceclass), в котором указываются конкретные параметры самих машин.

## Параметры
Настройки модуля устанавливаются автоматически на основании выбранной схемы размещения. В большинстве случаев нет необходимости в ручной конфигурации модуля.

Если вам необходимо настроить модуль, потому что, например, у вас bare metal кластер, для которого нужно включить
возможность добавлять дополнительные инстансы из OpenStack, то смотрите раздел как [настроить Hybrid кластер в OpenStack](usage.html#создание-гибридного-кластера).

Если у вас в кластере есть инстансы, для которых будут использоваться External Networks, кроме указанных в схеме размещения,
то их следует передавать в параметре

* `additionalExternalNetworkNames` — имена дополнительных сетей, которые могут быть подключены к виртуальной машине, и используемые `cloud-controller-manager` для проставления `ExternalIP` в `.status.addresses` в Node API объект.
    * Формат — массив строк.
      ```yaml
      cloudProviderOpenstack: |
        additionalExternalNetworkNames:
        - some-bgp-network
      ```

## Настройка политик безопасности на узлах кластера

Вариантов, зачем может понадобиться ограничить или наоборот расширить входящий или исходящий трафик на виртуальных
машинах кластера, может быть множество, например:
* Разрешить подключение к нодам кластера с виртуальных машин из другой подсети
* Разрешить подключение к портам статической ноды для работы приложения
* Ограничить доступ к внешним ресурсам или другим вм в облаке по требования службу безопасности

Для всего этого следует применять дополнительные security groups. Можно использовать только security groups, предварительно
созданные в облаке.

## Установка дополнительных security groups на мастерах и статических нодах
Данный параметр можно задать либо при создании кластера, либо в уже существующем кластере. В обоих случаях дополнительные
security groups указываются в `OpenStackClusterConfiguration`:
- Для мастеров — в секции `masterNodeGroup` в поле `additionalSecurityGroups`.
- Для статических нод — в секции `nodeGroups` в конфигурации, описывающей желаемую nodeGroup, также в поле `additionalSecurityGroups`.

Поле `additionalSecurityGroups` представляет собой массив строк с именами security groups.

## Установка дополнительных security groups на эфемерных нодах
Необходимо прописать параметр `additionalSecurityGroups` для всех OpenStackInstanceClass в кластере, которым нужны дополнительные
security groups. Смотри [параметры модуля cloud-provider-openstack](/modules/030-cloud-provider-openstack/configuration.html).

## Hybrid кластер в OpenStack

Hybrid кластер представляет собой объединённые в один кластер bare metal ноды и ноды openstack. Для создания такого кластера
необходимо наличие L2 сети между всеми нодами кластера.

### Параметры конфигурации

> **Внимание!** При изменении конфигурационных параметров приведенных в этой секции (параметров, указываемых в ConfigMap deckhouse) **перекат существующих Machines НЕ производится** (новые Machines будут создаваться с новыми параметрами). Перекат происходит только при изменении параметров `NodeGroup` и `OpenStackInstanceClass`. См. подробнее в документации модуля [node-manager](/modules/040-node-manager/faq.html#как-перекатить-эфемерные-машины-в-облаке-с-новой-конфигурацией).
Для настройки аутентификации с помощью модуля `user-authn` необходимо в Crowd'е проекта создать новое `Generic` приложение.

* `connection` - Параметры подключения к api cloud provider'a
    * `authURL` — OpenStack Identity API URL.
    * `caCert` — если OpenStack API имеет self-signed сертификат, можно указать CA x509 сертификат, использовавшийся для подписи.
        * Формат — строка. Сертификат в PEM формате.
        * Опциональный параметр.
    * `domainName` — имя домена.
    * `tenantName` — имя проекта.
        * Не может использоваться вместе с `tenantID`.
    * `tenantID` — id проекта.
        * Не может использоваться вместе с `tenantName`.
    * `username` — имя пользователя с полными правами на проект.
    * `password` — пароль к пользователю.
    * `region` — регион OpenStack, где будет развёрнут кластер.
* `internalNetworkNames` — имена сетей, подключённые к виртуальной машине, и используемые cloud-controller-manager для проставления InternalIP в `.status.addresses` в Node API объект.
    * Формат — массив строк. Например,

        ```yaml
        internalNetworkNames:
        - KUBE-3
        - devops-internal
        ```
* `externalNetworkNames` — имена сетей, подключённые к виртуальной машине, и используемые cloud-controller-manager для проставления ExternalIP в `.status.addresses` в Node API объект.
    * Формат — массив строк. Например,

        ```yaml
        externalNetworkNames:
        - KUBE-3
        - devops-internal
        ```
* `podNetworkMode` - определяет способ организации трафика в той сети, которая используется для коммуникации между подами (обычно это internal сеть, но бывают исключения).
    * Допустимые значение:
      * `DirectRouting` – между узлами работает прямая маршрутизация.
      * `DirectRoutingWithPortSecurityEnabled` - между узлами работает прямая маршрутизация, но только если в OpenStack явно разрешить на Port'ах диапазон адресов используемых во внутренней сети.
          * **Внимание!** Убедитесь, что у `username` есть доступ на редактирование AllowedAddressPairs на Port'ах, подключенных в сеть `internalNetworkName`. Обычно, в OpenStack, такого доступа нет, если сеть имеет флаг `shared`.
      * `VXLAN` – между узлами НЕ работает прямая маршрутизация, необходимо использовать VXLAN.
    * Опциональный параметр. По-умолчанию `DirectRoutingWithPortSecurityEnabled`.
* `instances` — параметры instances, которые используются при создании виртуальных машин:
    * `sshKeyPairName` — имя OpenStack ресурса `keypair`, который будет использоваться при заказе instances.
        * Обязательный парамер.
        * Формат — строкa.
    * `securityGroups` — Список securityGroups, которые нужно прикрепить к заказанным instances. Используется для задания firewall правил по отношению к заказываемым instances.
        * Опциональный параметр.
        * Формат — массив строк.
    * `imageName` - имя образа.
        * Опциональный параметр.
        * Формат — строкa.
    * `mainNetwork` - путь до network, которая будет подключена к виртуальной машине, как основная сеть (шлюз по умолчанию).
        * Опциональный параметр.
        * Формат — строкa.
    * `additionalNetworks` - список сетей, которые будут подключены к инстансу.
        * Опциональный параметр.
        * Формат — массив строк.
* `loadBalancer` - параметры Load Balancer
    * `subnetID` - ID Neutron subnet, в котором создать load balancer virtual IP.
        * Формат — строка.
        * Опциональный параметр.
    * `floatingNetworkID` - ID external network, который будет использоваться для заказа floating ip
        * Формат — строка.
        * Опциональный параметр.
* `zones` - список зон, в котором по умолчанию заказывать инстансы. Может быть переопределён индивидуально для каждой NodeGroup'ы
    * Формат — массив строк.
* `tags` - словарь тегов, которые будут на всех заказываемых инстансах
    * Опциональный параметр.
    * Формат — ключ-значение.