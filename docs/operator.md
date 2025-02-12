# Operator build and deployment

## Prerequisites

1. operator-sdk <https://sdk.operatorframework.io/docs/installation/>
2. kustomize <https://github.com/kubernetes-sigs/kustomize/releases>
3. opm <https://github.com/operator-framework/operator-registry/releases>

## Building the operator bundle (optional)

For development and testing purposes it may be beneficial to build the operator
bundle and index images. If you don't __need__ to build it, just skip to
[Deploying the Operator](#deploying-the-operator).

Build the bundle:

```
export BUNDLE_IMAGE=quay.io/${QUAY_NAMESPACE}/assisted-service-operator-bundle:${TAG}
make operator-bundle-build
```

## Deploying the operator

The operator must be deployed to the assisted-installer namespace. Create the namespace.

```bash
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Namespace
metadata:
  name: assisted-installer
  labels:
    name: assisted-installer
EOF
```

Having the ClusterDeployment CRD installed is a prerequisite.
Install Hive, if it has not already been installed.

``` bash
cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: hive-operator
  namespace: openshift-operators
spec:
  channel: alpha
  installPlanApproval: Automatic
  name: hive-operator
  source: community-operators
  sourceNamespace: openshift-marketplace
```

Deploy the operator using the operator-sdk:

```bash
operator-sdk run bundle \
  --namespace assisted-installer \
  ${BUNDLE_IMAGE:-quay.io/ocpmetal/assisted-service-operator-bundle:latest}
```

Now you should see the `assisted-service-operator` deployment running in the
`assisted-installer` namespace.

**NOTE**

Use `operator-sdk cleanup assisted-service-operator` to remove the operator when
installed via `operator-sdk run`.

## Creating an AgentServiceConfig Resource

The Assisted Service is deployed by creating an AgentServiceConfig.
At a minimum, you must specify the `databaseStorage` and `filesystemStorage` to
be used.


``` bash
cat <<EOF | kubectl create -f -
apiVersion: agent-install.openshift.io/v1beta1
kind: AgentServiceConfig
metadata:
  name: agent
spec:
  databaseStorage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  filesystemStorage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
EOF
```

## Configuring the Assisted Service Deployment

### Via Subscription

The operator subscription can be used to configure the images used in the
assisted-service deployment and the installer + controller + agent images used by
the assisted-service.

``` bash
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: assisted-service-operator
  namespace: assisted-installer
spec:
  config:
    env:
      - name: SERVICE_IMAGE
        value: ${SERVICE_IMAGE}
      - name: DATABASE_IMAGE
        value: ${DATABASE_IMAGE}
      - name: AGENT_IMAGE
        value: ${AGENT_IMAGE}
      - name: CONTROLLER_IMAGE
        value: ${CONTROLLER_IMAGE}
      - name: INSTALLER_IMAGE
        value: ${INSTALLER_IMAGE}
EOF
```

**NOTE**

The default channel for the assisted-service-operator package, here and in
[community-operators](https://github.com/operator-framework/community-operators/tree/master/community-operators/assisted-service-operator),
is `"alpha"` so we do not include it in the Subscription.

### Specifying Environmental Variables via ConfigMap

It is possible to specify a ConfigMap to be mounted into the assisted-service
container as environment variables by adding an
`"unsupported.agent-install.openshift.io/assisted-service-configmap"`
annotation to the `AgentServiceConfig` specifying the name of the configmap to be
used. This ConfigMap must exist in the namespace where the operator is
installed.

Simply create a ConfigMap in the `assisted-installer` namespace:

``` bash
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-assisted-service-config
  namespace: assisted-installer
data:
  LOG_LEVEL: "debug"
EOF
```

Add the annotation to the AgentServiceConfig:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: AgentServiceConfig
metadata:
  name: agent
  annotations:
    unsupported.agent-install.openshift.io/assisted-service-configmap: "my-assisted-service-config"
EOF
```

### Mirror Registry Configuration

In case user wants to create an installation that uses mirror registry, the following must be executed:
1. Create and upload a ConfigMap that will include the following keys:
   ca-bundle.crt -  contents of the certificate for accessing the mirror registry. I may be a certificate bundle, but it is still one string
   registries.conf - contents of the registries.conf file that configures mapping to the mirror registry.

``` bash
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: mirror-registry-ca
  namespace: "assisted-installer"
  labels:
    app: assisted-service
data:
  ca-bundle.crt: |
    -----BEGIN CERTIFICATE-----
    certificate contents
    -----END CERTIFICATE-----

  registries.conf: |
    unqualified-search-registries = ["registry.access.redhat.com", "docker.io"]
    
    [[registry]]
       prefix = ""
       location = "quay.io/ocpmetal"
       mirror-by-digest-only = false
    
       [[registry.mirror]]
       location = "bm-cluster-1-hyper.e2e.bos.redhat.com:5000/ocpmetal"
EOF
```

   Note1: ConfigMap should be installed in the same namespace as the assisted-service-operator (ie. `assisted-installer`).
   
   Note2: registry.conf supplied should use "mirror-by-digest-only = false" mode

2. set the mirrorRegistryConfigmapName in the spec of AgentServiceConfig to the name of uploaded ConfigMap. Example:

``` bash
cat <<EOF | kubectl apply -f -
apiVersion: agent-install.openshift.io/v1beta1
kind: AgentServiceConfig
metadata:
  name: agent
spec:
  mirrorRegistryConfigmapName: mirrorRegistyMap
EOF
```

For more details on how to specify the CR, see [AgentServiceConfig CRD](https://github.com/openshift/assisted-service/blob/master/config/crd/bases/agent-install.openshift.io_agentserviceconfigs.yaml).
