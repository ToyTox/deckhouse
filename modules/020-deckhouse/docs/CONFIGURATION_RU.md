---
title: "Модуль deckhouse: настройки"
---

<!-- SCHEMA -->

**Внимание!** В случае, если в `nodeSelector` указан несуществующий label, или указаны не верные `tolerations`, Deckhouse перестанет работать. Для восстановления работоспособности необходимо изменить значения на правильные в `configmap/deckhouse` и в `deployment/deckhouse`.
