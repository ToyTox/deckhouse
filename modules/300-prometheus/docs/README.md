---
title: "Модуль prometheus"
tags:
  - prometheus
type:
  - instruction
search: prometheus
permalink: /modules/300-prometheus/
sidebar: modules-prometheus
hide_sidebar: false
---

Модуль устанавливает [prometheus](https://prometheus.io/) (используя модуль [operator-prometheus](../200-operator-prometheus/)) и полностью его настраивает!

Модуль устанавливает два экземпляра Prometheus:
* **main** — основной Prometheus, который выполняет scrape каждые 30 секунд (с помощью параметра `scrapeInterval` можно изменить это значение). Именно он обрабатывает все правила, шлет алерты и является основным источником данных.
* **longterm** — дополнительный Prometheus, который выполняет scrape данных из main каждые 5 минут (с помощью параметра `longtermScrapeInterval` можно изменить это значение).. Используется для продолжительного хранения истории и для отображения больших промежутков времени.

Когда данный модуль установлен на Kubernetes 1.11 и имеет storage class, поддерживающий [автоматическое расширение диска](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/), в случае, если в prometheus не будет помещаться данные, то диск будет автоматически расширяться. В ином случае придет алерт о том, что дисковое пространство в Prometheus заканчивается.

Ресурсы cpu и memory автоматически выставляются при пересоздании пода на основе истории потребления, благодаря модулю [Vertical Pod Autoscaler](../302-vertical-pod-autoscaler). Так же, благодаря кешированию запросов к Prometheus с помощью trickster, потребление памяти Prometheus'ом сильно сокращается.

Дополнительная информация
-------------------------

* [Мониторинг приложений в проекте]({{ site.baseurl }}/guides/monitoring.html)
* [Разработка правил для Prometheus (алертов и recording)](prometheus_rules_development.html)
* [Разработка графиков (Dashboard'ов для Grafana)](grafana_dashboard_development.html)
* [Разработка target'ов для Prometheus (целей, которые мониторить)](prometheus_targets_development.html)

Конфигурация
------------

### Что нужно настраивать?

При установке можно ничего не настраивать.

### Параметры

* `retentionDays` — сколько дней хранить данные.
    * По-умолчанию `15`.
* `storageClass` — имя storageClass'а, который использовать.
    * Если не указано — используется StorageClass существующей PVC Prometheus, а если PVC пока нет — используется или `global.storageClass`, или `global.discovery.defaultStorageClass`, а если и их нет — данные сохраняются в emptyDir.
    * **ОСТОРОЖНО!** При указании этой опции в значение, отличное от текущего (из cуществующей PVC), диск Prometheus будет перезаказан, а все данные удалены.
* `longtermStorageClass` — имя storageClass'а, который использовать для Longterm Prometheus.
    * Если не указано — используется StorageClass существующей PVC Longterm Prometheus, а если PVC пока нет — используется или `prometheus.storageClass` от основного Prometheus, или `global.storageClass`, или `global.discovery.defaultStorageClass`, а если и их нет — данные сохраняются в emptyDir.
    * **ОСТОРОЖНО!** При указании этой опции в значение, отличное от текущего (из cуществующей PVC), диск Longterm Prometheus будет перезаказан, а все данные удалены.
* `longtermRetentionDays` — сколько дней хранить данные в longterm Prometheus.
    * По-умолчанию `1095`.
    * Если указать `0`, то longterm Prometheus не будет запущен в кластере.
* `auth` — опции, связанные с аутентификацией или авторизацией в приложении:
    * `externalAuthentication` - параметры для подключения внешней аутентификации (используется механизм Nginx Ingress [external-auth](https://kubernetes.github.io/ingress-nginx/examples/auth/external-auth/), работающей на основе модуля Nginx [auth_request](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html).
        * `authURL` - URL сервиса аутентификации. Если пользователь прошел аутентификацию, сервис должен возвращать код ответа HTTP 200.
        * `authSignInURL` - URL, куда будет перенаправлен пользователь для прохождения аутентификации (если сервис аутентификации вернул код ответа HTTP отличный от 200).
    * `password` — пароль для http-авторизации для пользователя `admin` (генерируется автоматически, но можно менять)
        * Используется если не включен параметр `externalAuthentication`.
    * `allowedUserGroups` — массив групп, пользователям которых позволен доступ в grafana и prometheus.
        * Используется если включен параметр `externalAuthentication` и модуль `user-authn`.
    * `whitelistSourceRanges` — массив CIDR, которым разрешено проходить авторизацию в grafana и prometheus.
    * `satisfyAny` — разрешает пройти только одну из аутентификаций. В комбинации с опцией whitelistSourceRanges позволяет считать авторизованными всех пользователей из указанных сетей без ввода логина и пароля.
* `grafana` - настройки для инсталляции Grafana.
    * `useDarkTheme` - использование по умолчанию пользовательской темной темы.
        * По-умолчанию `false`.
    * `customPlugins` - список дополнительных [plug-in'ов](https://grafana.com/grafana/plugins) для Grafana. Необходимо указать в качестве значения список имен плагинов из официального репозитория.
        * Пример добавления plug-in'ов для возможности указания в качестве datasource clickhouse и панели flow-chart:
           ```yaml
           grafana:
             customPlugins:
             - agenty-flowcharting-panel
             - vertamedia-clickhouse-datasource
           ```
* `ingressClass` — класс ingress контроллера, который используется для grafana/prometheus.
    * Опциональный параметр, по-умолчанию используется глобальное значение `modules.ingressClass`.
* `https` — выбираем, какой типа сертификата использовать для grafana/prometheus.
    * При использовании этого параметра полностью переопределяются глобальные настройки `global.modules.https`.
    * `mode` — режим работы HTTPS:
        * `Disabled` — в данном режиме grafana/prometheus будут работать только по http;
        * `CertManager` — grafana/prometheus будут работать по https и заказывать сертификат с помощью clusterissuer заданном в параметре `certManager.clusterIssuerName`;
        * `CustomCertificate` — grafana/prometheus будут работать по https используя сертификат из namespace `d8-system`;
        * `OnlyInURI` — grafana/prometheus будет работать по http (подразумевая, что перед ними стоит внешний https балансер, который терминирует https) и все ссылки в `user-authn` будут генерироваться с https схемой.
    * `certManager`
      * `clusterIssuerName` — указываем, какой ClusterIssuer использовать для grafana/prometheus (в данный момент доступны `letsencrypt`, `letsencrypt-staging`, `selfsigned`, но вы можете определить свои).
        * По-умолчанию `letsencrypt`.
    * `customCertificate`
      * `secretName` - указываем имя secret'а в namespace `d8-system`, который будет использоваться для grafana/prometheus (данный секрет должен быть в формате [kubernetes.io/tls](https://kubernetes.github.io/ingress-nginx/user-guide/tls/#tls-secrets)).
        * По-умолчанию `false`.
* `vpa`
    * `maxCPU` — максимальная граница CPU requests, выставляемая VPA контроллером для pod'ов основного Prometheus.
        * Значение по-умолчанию подбирается автоматически, исходя из максимального количества подов, которое можно создать в кластере при текущем количестве узлов и их настройках. Подробнее см. хук `detect_vpa_max` модуля.
    * `maxMemory` — максимальная граница Memory requests, выставляемая VPA контроллером для pod'ов основного Prometheus.
        * Значение по-умолчанию подбирается автоматически, исходя из максимального количества подов, которое можно создать в кластере при текущем количестве узлов и их настройках. Подробнее см. хук `detect_vpa_max` модуля.
    * `longtermMaxCPU` — максимальная граница CPU requests, выставляемая VPA контроллером для pod'ов longterm Prometheus.
        * Значение по-умолчанию подбирается автоматически, исходя из максимального количества подов, которое можно создать в кластере при текущем количестве узлов и их настройках. Подробнее см. хук `detect_vpa_max` модуля.
    * `longtermMaxMemory` — максимальная граница Memory requests, выставляемая VPA контроллером для pod'ов longterm Prometheus.
        * Значение по-умолчанию подбирается автоматически, исходя из максимального количества подов, которое можно создать в кластере при текущем количестве узлов и их настройках. Подробнее см. хук `detect_vpa_max` модуля.
    * `updateMode` — режим обновления Pod'ов.
        * По-умолчанию `Initial`, но возможно поставить `Auto` или `Off`.
* `highAvailability` — ручное управление [режимом отказоустойчивости]({{ site.baseurl }}/features.html#отказоустойчивость).
* `scrapeInterval` — с помощью данного параметра можно указать, как часто prometheus будет собирать метрики с таргетов. Evaluation Interval всегда равен scrapeInterval.
    * По-умолчанию `30s`.
* `longtermScrapeInterval` — с помощью данного параметра можно указать, как часто longterm prometheus будет собирать себе "снимок" данных из основного prometheus.
    * По-умолчанию `5m`.
* `nodeSelector` — как в Kubernetes в `spec.nodeSelector` у pod'ов.
    * Если ничего не указано — будет [использоваться автоматика]({{ site.baseurl }}/#выделение-узлов-под-определенный-вид-нагрузки).
    * Можно указать `false`, чтобы не добавлять никакой nodeSelector.
* `tolerations` — как в Kubernetes в `spec.tolerations` у pod'ов.
    * Если ничего не указано — будет [использоваться автоматика]({{ site.baseurl }}/#выделение-узлов-под-определенный-вид-нагрузки).
    * Можно указать `false`, чтобы не добавлять никакие toleration'ы.
* `mainMaxDiskSizeGigabytes` — максимальный размер в гигабайтах, до которого автоматически может ресайзиться диск Prometheus main.
    *  Опциональный параметр, значение по-умолчанию — `300`.
* `longtermMaxDiskSizeGigabytes` — максимальный размер в гигабайтах, до которого автоматически может ресайзиться диск Prometheus longterm.
    *  Опциональный параметр, значение по-умолчанию — `300`.

### Примечание
* `retentionSize` для `main` и `longterm` **рассчитывается автоматически, возможности задать значение нет!**
    * Алгоритм расчета:
        * `pvc_size * 0.8` — если PVC существует.
        * `10 GiB` — если PVC нет и StorageClass поддерживает ресайз.
        * `25 GiB` — если PVC нет и StorageClass не поддерживает ресайз.
    * Если используется `local-storage` и требуется изменить `retentionSize`, то необходимо вручную изменить размер PV и PVC в нужную сторону. **Внимание!** Для расчета берется значение из `.status.capacity.storage` PVC, поскольку оно отражает рельный размер PV в случае ручного ресайза.

### Пример конфига

```yaml
prometheus: |
  password: xxxxxx
  retentionDays: 7
  storageClass: rbd
  nodeSelector:
    node-role/example: ""
  tolerations:
  - key: dedicated
    operator: Equal
    value: example
```
### Dashboard'ы для Grafana
Для хранения dashboard'ов добавлен специальный ресурс - `GrafanaDashboardDefinition`. [Читайте подробнее](grafana_dashboard_development.html) в документации по разработке графиков Grafana.

Пример:
```yaml
apiVersion: deckhouse.io/v1alpha1
kind: GrafanaDashboardDefinition
metadata:
  name: my-dashboard
spec:
  folder: My folder # Папка, в которой в Grafana будет отображаться ваш dashboard
  definition: |
    {
      "annotations": {
        "list": [
          {
            "builtIn": 1,
...
```
### Rule'ы для Prometheus
Для хранения PrometheusRule добавлен специальный ресурс - `CustomPrometheusRules`. [Читайте подробнее](../../guides/monitoring.html) в документации по добавлению пользовательских алертов.

Пример:
```yaml
apiVersion: deckhouse.io/v1alpha1
kind: CustomPrometheusRules
metadata:
  name: my-rules
spec:
  groups:
  - name: cluster-state-alert.rules
    rules:
    - alert: CephClusterErrorState
      annotations:
        description: Storage cluster is in error state for more than 10m.
        summary: Storage cluster is in error state
        plk_markup_format: markdown
        plk_protocol_version: "1"
      expr: |
        ceph_health_status{job="rook-ceph-mgr"} > 1
```

### Дополнительные Datasource для Grafana
Для подключения дополнительных datasource'ов к Grafana добавлен специальный ресурс - `GrafanaAdditionalDatasource`.

Параметры ресурса подробно описаны в [документации к Grafana](https://grafana.com/docs/grafana/latest/administration/provisioning/#example-datasource-config-file). 

Пример:
```yaml
apiVersion: deckhouse.io/v1alpha1
kind: GrafanaAdditionalDatasource
metadata:
  name: another-prometheus
spec:
  type: prometheus
  access: direct
  url: http://another-prometheus.example.com/prometheus
  jsonData:
    timeInterval: 30s
```