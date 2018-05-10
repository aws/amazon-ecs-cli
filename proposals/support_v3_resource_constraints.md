# Support container CPU and Memory constraints for Compose 3

## Introduction

### Background
Today, the ECS CLI supports the `cpu_shares`, `memory_limit` and `memory_reservation` resource contraint options in Compose file version 2 and older. When running `ecs-cli compose create` or `ecs-cli compose up`, these options are translated into equivalent [containerDefinition](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_ContainerDefinition.html) values for use within an ECS taskDefinition. These constraints are used by ECS to [find and place a task](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-placement.html) on an instance with sufficient resources.    

### Problem to solve
In [Compose file version 3](https://docs.docker.com/compose/compose-file/), resource constraints were moved into the [`deploy.resources`](https://docs.docker.com/compose/compose-file/#resources) option for use with docker swarm. This presents a problem for developers who wish to apply resource constraints to their containers outside of swarm ([see discussion](https://github.com/docker/compose/issues/4513)), and specifically for ECS CLI users who choose to use Compose 3, because container*-level memory limit and/or reservation is required to register a non-FARGATE task definition with ECS.

In order for the ECS CLI to support the Compose 3 file format (see [issue #218](https://github.com/aws/amazon-ecs-cli/issues/218)), we need a way to add support for container-level resource constraints that is equivalent to what users can do today with previous Compose versions.

*analogous to a docker compose service in the ECS CLI context


## Option 1: Support `deploy.resources` CPU & Memory values
This approach requires customers to define desired `cpu` and `memory` resources under the `deploy.resources` key in order to set container-level resource constraints in ECS. The ECS CLI would read these values from the docker ServiceConfig and port them into the appropriate field for the resulting ECS task definition. The proposed mapping would be:

* `deploy.resources.limits.memory` --> `[containerDefinition].memory`
* `deploy.resources.reservations.memory` --> `[containerDefinition].memoryReservation`
* `deploy.resources.reservations.cpu` --> `[containerDefinition].cpu`

The main benefit of this approach is that resource limits will continue to be defined within the *docker-compose.yml* file as they are today, however other options under `deploy` will not be supported.

### Pros
* Service settings will be contained within the **docker-compose.yml** file.
* `deploy.limits.memory` and `deploy.reservations.memory` fields are conceptually equivalent between docker swarm and ECS.

### Cons
* Only a subset of possible `deploy` CPU values can be supported in ECS. 
    * Percentage-based CPU allocation (see examples [here](https://docs.docker.com/get-started/part5/#persist-the-data) and [here](https://docs.docker.com/compose/compose-file/#resources)), is supported in Compose 3 `deploy.resources` CPU, but not in ECS. [`[containerDefinition].cpu`](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_ContainerDefinition.html) must be and integer, as it corresponds to the docker run [`cpu_shares`](https://docs.docker.com/engine/reference/run/#cpu-share-constraint) option.
    * Numberic values for `deploy.resources...cpu` also present a problem because they have differnt possible values in `deploy` vs. ECS. In ECS, `cpu` directly corresponds to `cpu_shares` and represents an INT of CPU units reserved for the container; but in the swarm context `cpu` is definied as a [DECIMAL](https://github.com/moby/moby/issues/28456#issuecomment-260810538), which means ECS CLI may need to round up/down to satisfy the `[containerDefinition].cpu` field and makes translation from Compose to ECS task less transparent.   
* Future changes or feature additions to `deploy` could complicate support of this field within the ECS CLI. For example, the key may be expended to contain sub-options or units (as memory does) under `cpu` (e.g., NanoCPUs), in which case porting to the corresponding ECS option may be impossible or require even more modification of the customer's provided value.
* Use of `deploy` with the ECS CLI would not prevent use of fields normally ignored in the docker swarm context, such as tmpfs, secuirty_opt, [and others](https://docs.docker.com/compose/compose-file/#not-supported-for-docker-stack-deploy). This introduces an inconsistency in the behavior of this option between docker/swarm and ECS CLI.


## Option 2: Add `cpu_shares`, `memory_limit` and `memory_reservation` service-level fields in *ecs-params.yml*
This approach avoids shoehorning in potentially incompatible values by allowing users to set the same resource values they can today through use of new fields in *ecs-params.yml*. No modification of user-entered values is required, although use of Compose 3 outside of FARGATE would now require configuration outside the *docker-compose.yml* file. 

**Proposed changes to ecs-params.yml:**
```yaml
version: 1
task_definition:
  ecs_network_mode: string
  task_role_arn: string
  task_execution_role: string
  task_size:
    cpu_limit: string
    mem_limit: string
  services:
    <service_name>:
      essential: boolean
      cpu_shares: number # NEW
      memory: number # NEW
      memory_reservation: number # NEW
run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: 
        - subnet_id1 
        - subnet_id2
      security_groups: 
        - secgroup_id1
        - secgroup_id2
      assign_public_ip: ENABLED
```

### Pros
* Resource constraints directly correspond to `docker run` values of the same name and their ECS equivalents.
* Decouple resource allocation from swarm syntax, future changes to `deploy.resource` fields.
* Less confusion re: what values are valid and actually enforced on the task; something [observed](https://github.com/moby/moby/issues/30222) in the swarm context.

### Cons
* Requires use of an **ecs-params.yml** file for any non-FARGATE v3 project, a potential adoption barrier for users less familiar with ECS constructs.
* More fields to maintain & update as application requirements evolve.


## Recommendation
**Option 2** is the better choice in terms of:
1. Transparency of what's being enforced re: CPU while the task is running.
2. Long-term support of resource constraints within the ECS CLI.

As the Compose file format continues to evolve alongside swarm, run-time options like resource constraints can be expected to further adhere to the particulars of that system's requirements. Decoupling ECS-relevant resource constraints from purpose-built swarm fields gives customers more control over their ECS tasks and let's them run containers with the same set of constraints they do today.

However, the need for an additional configuration (*ecs-params.yml*) outside of ECS-specific features is unfortunate. Implemention of this option should entail clear messaging/warnings when required fields are missing or found under `deploy` to minimize confusion for customers who adopt Compose 3. We should be prepared to iterate on this solution as customers use the new format and submit feedback.







