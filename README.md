# guacamole
A High Availability manager and API for postgres

## Introduction
guacamole is just the project name for a personal project on Postgres clutser mnagement exposed with an API.

It aims at 2 distinct functionalities:
1. Cluster managament with state of the art technologies
2. Being controlplane for your postgres, exposed as an API

As a sidenote, we will try to allow running guacamole as a sidecar rather thn as part of the main Postgres container. This seems achievable, but we still need to find a way that works for this...

This means that when guacamole is running on your cluster servers / pods Postgres (by fefault) is not initialized, running, started, etc.
But the controlplane is, and it alows you to create a cluster anyway you like.
In the future, we might add a init option, which allows you to start it automated without requiring any interaction.
That being said, still guacamole is there as a contollane for your postgres (where the postgres port will remain the dataplane).

Next to that:
* it allows me to train myself better in conceptual thinking
* it allows me to train programing skills
* it allows me to demonstrate 'higher level thinking' as a concept

### Whats in a name
The name is a project name. Probably this will become another name if this project matures beyond a certain level.
guacamole has a nice sound, a fresh touch to it, etc.

Furthermore, (to a dutch guy), it sounds a bit like wack-a-mole', the game where you ht the mole and directly when it is down, another pops up.
You do not know which, but another one always pops up. This (always fast recovery from master failure) is what HA should look like for Postgres.

## Current design
guacamole is built with these components:
* a cluster component, which changes separate instances into a cluster. Basis is Raft, which manages leader election. Leader:
  * is SPOC for config
  * elects the new master in a failover situation
  * is the endpoint for all API calls (ending up on any member)
* an API (designed with swagger), which allows for all controlplane functionalities, like
  * get / change topology (includes async/sync and logical/physical replicaton, configuring upstream, etc.)
  * configure access control (pg_hba, user management, break the glass functionality)
  * initialize
  * enforce postgres configuration
* a Postgres Manager, which allows to
  * initialize the cluster (initdb)
  * clone
  * create / change config files
  * start, stop, restart, reload  Postgres
  * etc.
* a postgres connection, which allows guacamole to check and enforce stuff within Postgres, aswell as monitor Postgres (only for HA purposes)

## development process

### changing code
At start of development this is a private project, and therefore starting a a private github repo.
I aim at developing it as a public repo, and hope for my current employer to adopt it as an Open Source project, which allows multiple users of the system.
As part of that dream, I will start as the one developer on this project and make it Apache 2.0 license.
Once I am at a point where I can invite other collaborators / become a public projecy, this will be the process:
1. fork on github
2. apply your changes
3. make sure you have unittests, meet code quality standards, etc.
  * Makefile will help you to get there
4. create a pull request

### design and Roadmap
1. design. This is a offline process (your own mind, mail, etc).
2. share. Put all your thoughts in a file in the `roadmap` folder, and create a merge request.
3. Merging means review and create the actual tickets for it.

## Answers to obvious questions
1. Why not patroni/stolon?
  * Patroni and stolon are awesome. They are two different approaches to the same problem. This aims at becoming a third.
  * guacamole might become a similar approach to stolon, and might be abandoned, because of that. I just find it interesting to approach it from this angle.
2. What is and is not part ogf the cluster?
  * currently this is about Streaming replication for Postgres. Physical, logical and sync/async replication is all part of that.
  * if bdr will ever be part of that depends greatly on it adoption
  * pooling, backup and other options are not planned to be part of this project directly, but
    * might consume guacamoles API for interaction
    * might be configured in hooks for interaction from guacamole
    * might be configured as part of this project to some extent (imaging archive_command used for backup solution, managed by guacamole)
    * locally running components might still be managed by guacamole (makes sense to have only one API that does all required for a propery running postgres).

