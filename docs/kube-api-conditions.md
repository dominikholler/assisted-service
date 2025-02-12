# Hive Integration - Conditions

Conditions provide a standard mechanism for higher-level status reporting from a controller. Read more about conditions [here](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)

## ClusterDeployment Conditions

Hive handle a set of existing condition [types](https://github.com/openshift/hive/blob/master/apis/hive/v1/clusterdeployment_types.go#L338).

Assisted Installer handles additional types: `SpecSynced`, `ReadyForInstallation`, `Installed` and `Validated`.

|Type|Status|Reason|Message|Description|
|----|----|-----|-------------------|-------------------|
|SpecSynced|True|SyncOK|The Spec has been successfully applied|If the Spec was successfully applied|
|SpecSynced|False|BackendError|The Spec could not be synced due to backend error: <err>|If the Spec was not applied due to 500 error|
|SpecSynced|False|InputError|The Spec could not be synced due to an input error: <err>|If the Spec was not applied due to 40X error|
||||||
|Validated|True|ValidationsPassing|The cluster's validations are passing|Otherwise than other conditions|
|Validated|False|ValidationsFailing|The cluster's validations are failing: "summary of failed validations"|If the cluster status is "insufficient" or "pending-for-input"|
|Validated|Unknown|ValidationsUnknown|The cluster's validations have not yet been calculated|If the validations have not yet been calculated|
||||||
|ReadyForInstallation|True|ClusterIsReady|The cluster is ready to begin the installation|if the cluster status is "ready"|
|ReadyForInstallation|False|ClusterNotReady|The cluster is not ready to begin the installation|If the cluster is before installation ("insufficient"/"pending-for-input")|
|ReadyForInstallation|False|ClusterAlreadyInstalling|The cluster cannot begin the installation because it has already started|If the cluster has begun installing ("preparing-for-installation", "installing", "finalizing", "installing-pending-user-action", "adding-hosts", "installed", "error") |
||||||
|Installed|True|InstallationCompleted|The installation has completed: "status_info"|If the cluster status is "installed"|
|Installed|False|InstallationFailed|The installation has failed: "status_info"|If the cluster status is "error"|
|Installed|False|InstallationNotStarted|The installation has not yet started|If the cluster is before installation ("insufficient"/"pending-for-input"/"ready")|
|Installed|False|InstallationInProgress|The installation is in progress: "status_info"|If the cluster is installing ("preparing-for-installation", "installing", "finalizing", "installing-pending-user-action")|

Here an example of ClusterDeployment conditions:

```sh
Status:
  Conditions:
    Last Probe Time:       2021-04-20T12:40:38Z
    Last Transition Time:  2021-04-20T12:40:38Z
    Message:               The Spec has been successfully applied
    Reason:                SyncOK
    Status:                True
    Type:                  SpecSynced
    Last Probe Time:       2021-04-20T12:52:58Z
    Last Transition Time:  2021-04-20T12:52:58Z
    Message:               The cluster cannot begin the installation because it has already started
    Reason:                ClusterAlreadyInstalling
    Status:                False
    Type:                  ReadyForInstallation
    Last Probe Time:       2021-04-20T12:47:01Z
    Last Transition Time:  2021-04-20T12:47:01Z
    Message:               The cluster s validations are passing
    Reason:                ValidationsPassing
    Status:                True
    Type:                  Validated
    Last Probe Time:       2021-04-20T12:52:58Z
    Last Transition Time:  2021-04-20T12:52:58Z
    Message:               The installation is in progress: Preparing cluster for installation
    Reason:                InstallationInProgress
    Status:                False
    Type:                  Installed
```

## Agent Conditions

The Agent condition types supported are: `SpecSynced`, `Connected`, `ReadyForInstallation`, `Validated` and `Installed`

|Type|Status|Reason|Message|Description|
|----|----|-----|-------------------|-------------------|
|SpecSynced|True|SyncOK|The Spec has been successfully applied|If the Spec was successfully applied|
|SpecSynced|False|BackendError|The Spec could not be synced due to backend error: <err>|If the Spec was not applied due to 500 error|
|SpecSynced|False|InputError|The Spec could not be synced due to an input error: <err>|If the Spec was not applied due to 40X error|
||||||
|Validated|True|ValidationsPassing|The agent's validations are passing|Otherwise than other conditions|
|Validated|False|ValidationsFailing|The agent's validations are failing: "summary of failed validations"|If the host status is "insufficient"|
|Validated|Unknown|ValidationsUnknown|The agent's validations have not yet been calculated|If the validations have not yet been calculated|
||||||
|ReadyForInstallation|True|AgentIsReady|The agent is ready to begin the installation|if the host status is "known"|
|ReadyForInstallation|False|AgentNotReady|The agent is not ready to begin the installation|If the host is before installation ("discovering"/"insufficient"/"disconnected"/"pending-input")|
|ReadyForInstallation|False|AgentAlreadyInstalling|The agent cannot begin the installation because it has already started|If the agent has begun installing ("preparing-successful","preparing-for-installation", "installing", "installed", "error") |
||||||
|Installed|True|InstallationCompleted|The installation has completed: "status_info"|If the host status is "installed"|
|Installed|False|InstallationFailed|The installation has failed: "status_info"|If the host status is "error"|
|Installed|False|InstallationNotStarted|The installation has not yet started|If the cluster is before installation ("discovering"/"insufficient"/"disconnected"/"pending-input/known")|
|Installed|False|InstallationInProgress|The installation is in progress: "status_info"|If the host is installing ("preparing-for-installation", "preparing-successful", "installing")|
||||||
|Connected|True|AgentIsConnected|The agent has not contacted the installation service in some time, user action should be taken|If the host status is not "disconnected"|
|Connected|False|AgentIsDisconnected|The agent's connection to the installation service is unimpaired|If the host status is "error"|


Here an example of Agent conditions:

```sh
Status:
  Conditions:
    Last Transition Time:  2021-04-22T15:50:24Z
    Message:               The Spec has been successfully applied
    Reason:                SyncOK
    Status:                True
    Type:                  SpecSynced
    Last Transition Time:  2021-04-22T15:50:24Z
    Message:               The agent's connection to the installation service is unimpaired
    Reason:                AgentIsConnected
    Status:                True
    Type:                  Connected
    Last Transition Time:  2021-04-22T15:50:33Z
    Message:               The agent cannot begin the installation because it has already started
    Reason:                AgentAlreadyInstalling
    Status:                False
    Type:                  ReadyForInstallation
    Last Transition Time:  2021-04-22T15:50:26Z
    Message:               The agent's validations are passing
    Reason:                ValidationsPassing
    Status:                True
    Type:                  Validated
    Last Transition Time:  2021-04-22T15:50:24Z
    Message:               The installation is in progress: Host is preparing for installation
    Reason:                InstallationInProgress
    Status:                False
    Type:                  Installed
```

## InfraEnv Conditions

The InfraEnv condition type supported is: `ImageCreated`

|Type|Status|Reason|Message|Description|
|----|----|-----|-------------------|-------------------|
|ImageCreated|True|ImageCreated|Image has been created|If the ISO image was successfully created|
|ImageCreated|False|ImageCreationError|Failed to create image: "error message"If the ISO image was not successfully created|

Here an example of InfraEnv conditions:

```sh
Status:
  Conditions:
    Last Transition Time:  2021-04-22T15:49:35Z
    Message:               Image has been created
    Reason:                ImageCreated
    Status:                True
    Type:                  ImageCreated
```
