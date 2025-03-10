---
title: "Сloud provider — AWS: примеры конфигурации"
---

## Пример CR `AWSInstanceClass`

```yaml
apiVersion: deckhouse.io/v1
kind: AWSInstanceClass
metadata:
  name: worker
spec:
  instanceType: t3.large
  ami: ami-040a1551f9c9d11ad
  diskSizeGb: 15
  diskType:  gp2
```

## LoadBalancer

### Аннотации объекта Service

Поддерживаются следующие параметры в дополнение к существующим в upstream:

1. `service.beta.kubernetes.io/aws-load-balancer-type` — может иметь значение `none`, что приведёт к созданию **только** Target Group, без какого-либо LoadBalanacer.
2. `service.beta.kubernetes.io/aws-load-balancer-backend-protocol` — используется в связке с `service.beta.kubernetes.io/aws-load-balancer-type: none`.
   * Возможные значения:
     * `tcp` (по умолчанию)
     * `tls`
     * `http`
     * `https`
   * **Внимание!** При изменении поля `cloud-controller-manager` попытается пересоздать Target Group. Если к ней уже привязаны NLB или ALB, удалить Target Group он не сможет и будет пытаться вечно. Необходимо вручную отсоединить от Target Group NLB или ALB.

## Настройка политик безопасности на узлах

Вариантов, зачем может понадобиться ограничить или наоборот расширить входящий или исходящий трафик на виртуальных машинах кластера в AWS, может быть множество. Например:

* Разрешить подключение к узлам кластера с виртуальных машин из другой подсети.
* Разрешить подключение к портам статического узла для работы приложения.
* Ограничить доступ к внешним ресурсам или другим виртуальным машинам в облаке по требованию службы безопасности.

Для всего этого следует применять дополнительные security groups. Можно использовать только предварительно созданные в облаке security groups.

## Установка дополнительных security groups на статических и master-узлах

Данный параметр можно задать либо при создании кластера, либо в уже существующем кластере. В обоих случаях дополнительные security groups указываются в `AWSClusterConfiguration`:
- для master-узлов — в секции `masterNodeGroup` в поле `additionalSecurityGroups`
- для статических узлов — в секции `nodeGroups` в конфигурации, описывающей соответствующую nodeGroup, в поле `additionalSecurityGroups`.

Поле `additionalSecurityGroups` — содержит массив строк с именами security groups.

## Установка дополнительных security groups на эфемерных узлах

Необходимо указать параметр `additionalSecurityGroups` для всех [`AWSInstanceClass`](cr.html#awsinstanceclass) в кластере, которым нужны дополнительные security groups.

## Настройка балансировщика в случае наличия Ingress-узлов не во всех зонах

Необходимо указать аннотацию на объекте Service: `service.beta.kubernetes.io/aws-load-balancer-subnets: subnet-foo, subnet-bar`.

Получить список текущих подсетей, используемых для конкретной инсталляции:

```bash
kubectl -n d8-system exec deploy/deckhouse -c deckhouse -- deckhouse-controller module values cloud-provider-aws -o json \
| jq -r '.cloudProviderAws.internal.zoneToSubnetIdMap'
```
