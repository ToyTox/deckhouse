---
title: "Настройка политик безопасности на нодах кластера в AWS" 
---

## Сценарий использования
Вариантов, зачем может понадобиться ограничить или наоборот расширить входящий или исходящий трафик на виртуальных
машинах кластера, может быть множество, например: 
* Разрешить подключение к нодам кластера с виртуальных машин из другой подсети
* Разрешить подключение к портам статической ноды для работы приложения
* Ограничить доступ к внешним ресурсам или другим вм в облаке по требованию службы безопасности

Для всего этого следует применять дополнительные security groups. Можно использовать только security groups, предварительно
созданные в облаке.

## Установка дополнительных security groups на мастерах и статических нодах
Данный параметр можно задать либо при создании кластера, либо в уже существующем кластере. В обоих случаях дополнительные
security groups указываются в AWSClusterConfiguration. Для мастеров в секции `masterNodeGroup` в поле `additionalSecurityGroups`,
для статических нод в секции `nodeGroups` в конфигурации, описывающей желаемую nodeGroup, также в поле `additionalSecurityGroups`.
Поле `additionalSecurityGroups` представляет собой массив строк с именами security groups.

## Установка дополнительных security groups на эфемерных нодах
Необходимо прописать параметр `additionalSecurityGroups` для всех AWSInstanceClass в кластере, которым нужны дополнительные
security groups. Смотри [параметры модуля cloud-provider-aws](/modules/030-cloud-provider-aws/).