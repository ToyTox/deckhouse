---
title: "Cloud provider — AWS: FAQ"
---

## How do I create a peering connection between VPCs?

Let's, for example, create a peering connection between two VPCs, vpc-a and vpc-b.

**Caution!**
IPv4 CIDR must be unique for each VPC.

* Switch to the region where vpc-a is running.
* VPC -> VPC Peering Connections -> Create Peering Connection, configure a peering connection:

  * Name: vpc-a-vpc-b
  * Fill in Local and Another VPC fields.

* Switch to the region where vpc-b is running.
* VPC -> VPC Peering Connections.
* Select the newly created perring connection and click Action "Accept Request".
* Add routes to vpc-b's CIDR over a peering connection to the vpc-a's routing tables.
* Add routes to vpc-a's CIDR over a peering connection to the vpc-b's routing tables.


## How do I create a cluster in a new VPC with access over an existing bastion host?

* Bootstrap the base-infrastructure of the cluster:

  ```shell
  dhctl bootstrap-phase base-infra --config config
  ```

* Set up a peering connection using the instructions [above](#how-do-i-create-a-peering-connection-between-vpcs).
* Continue installing the cluster, enter "y" when asked about the terraform cache:

  ```shell
  dhctl bootstrap --config config --ssh-...
  ```

## How do I create a cluster in a new VPC and set up bastion host to access the nodes?

* Bootstrap the base-infrastructure of the cluster:

  ```shell
  dhctl bootstrap-phase base-infra --config config
  ```

* Manually set up the bastion host in the subnet <prefix>-public-0.
* Continue installing the cluster, enter "y" when asked about the terraform cache:

  ```shell
  dhctl bootstrap --config config --ssh-...
  ```

## Configuring a bastion host

There are two possible cases:
* a bastion host already exists in an external VPC; in this case, you need to:
  * Create a basic infrastructure: `dhctl bootstrap-phase base-infra`;
  * Set up peering connection between an external and a newly created VPC;
  * Continue the installation by specifying the bastion: `dhctl bootstrap --ssh-bastion...`
* a bastion host needs to be deployed to a newly created VPC; in this case, you need to:
  * Create a basic infrastructure: `dhctl bootstrap-phase base-infra`;
  * Manually run a bastion in the <prefix>-public-0 subnet;
  * Continue the installation by specifying the bastion: `dhctl bootstrap --ssh-bastion...`

## Adding CloudStatic nodes to the cluster

To add a pre-created instance to the cluster, you need:
  * Attach a security group `<prefix>-node`
  * Add tags:

  ```
  "kubernetes.io/cluster/<cluster_uuid>" = "shared"
  "kubernetes.io/cluster/<prefix>" = "shared"
  ```

    * You can find out the `cluster_uuid` using the command:

      ```shell
      kubectl -n kube-system get cm d8-cluster-uuid -o json | jq -r '.data."cluster-uuid"'
      ```

    * You can find out `prefix` using the command:

      ```shell
      kubectl -n kube-system get secret d8-cluster-configuration -o json | jq -r '.data."cluster-configuration.yaml"' | base64 -d | grep prefix
      ```

## How to increase the size of a volume?

* Set the new size in the corresponding PersistentVolumeClaim resource, in the `spec.resources.requests.storage` parameter.
* The progress of the process can be observed in events using the command `kubectl describe pvc`.
* The operation is fully automatic and takes up to one minute. No further action is required.

> ℹ️ After modifying a volume, you must wait at least six hours and ensure that the volume is in the in-use or available state before you can modify the same volume. This is sometimes referred to as a cooldown period. You can find details in the [official documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/modify-volume-requirements.html).
