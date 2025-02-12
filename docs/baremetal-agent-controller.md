Baremetal Agent Controller (a.k.a BMAC)
==

BMAC is a Kubernetes controller responsible for reconciling [BareMetalHost][bmo] and Agent (defined
and maintained in this repo) resources for the agent-based deployment scenario.

Testing
==

The testing environment for BMAC consists in:

- [Downstream dev-scripts](https://github.com/openshift-metal3/dev-scripts/) deployment
- [Baremetal Operator][bmo]: It defines the BareMetalHost custom resource
- [Assisted Installer Operator](./operator.md): To deploy and manage the  assisted installer
  deployment. Read the operator docs to know more about its dependencies and installation process.


Each of the components listed above provide their own documentation on how to deploy and configure
them. However, you can find below a set of recommended configs that can be used for each of these
components:

Dev Scripts
===

```bash
# Giving Master nodes some extra CPU since we won't be
# deploying any workers
export MASTER_VCPU=4
export MASTER_MEMORY=20000

# No workers are needed to test BMAC
export NUM_WORKERS=0

# Add extra workers so we can use it for the deployment
export NUM_EXTRA_WORKERS=1

# At the time of this writing, this requires the 1195 PR
# mentioned below.
export PROVISIONING_NETWORK_PROFILE=Disabled

# By default dev-scripts uses the metal3 BMO image
# clone the downstream version and set the path here.
export BAREMETAL_OPERATOR_LOCAL_IMAGE=/path/to/your/local/clone
```

The config above should provide you with an environment that is ready to be used for the operator,
assisted installer, and BMAC tests. Here are a few tips that would help simplifying the environment
and the steps required:

- Clone [baremetal-operator][bmo] somewhere and set the BAREMETAL_OPERATOR_LOCAL_IMAGE in your config.

Once dev-script is up and running, modify the worker(s) and add 2 more disks (10GB should be enough)
as they are required by the Assisted Installer Operator.

Baremetal Operator
==

The [baremetal-operator][bmo] will define the BareMetalHost custom resource required by the agent
based install process. Setting the `BAREMETAL_OPERATOR_LOCAL_IMAGE` should build and run the BMO
already. However, it's recommended to run the [local-bmo][local-bmo] script to facilitated the
deployment and monitoring of the BMO. Here's what using [local-bmo][local-bmo] looks like:

It's possible to disable inspection for the master (and workers) nodes before running the local-bmo
script. This will make the script less noisy which will make debugging easier.

```bash
./metal3-dev/pause-control-plane.sh
```
The pause-control-plane script only pauses the control plane. You can do the same for the worker
nodes with the following command

```bash
for host in $(oc get baremetalhost -n openshift-machine-api -o name | grep -e '-worker-'); do
    oc annotate --overwrite -n openshift-machine-api "$host" \
       'baremetalhost.metal3.io/paused=""'
done
```

The steps mentioned above are optional, but recommended. They will make debugging easier. Let's now
run [local-bmo][local-bmo] and move on. This script will tail the logs so do it in a separate buffer
so that it can be kept running.

```bash
# Note variable is different from the one in your dev-script
# config file. You can set it to the same path, though.
export BAREMETAL_OPERATOR_PATH=/path/to/your/local/clone
./metal3-dev/local-bmo.sh

```

Assisted Installer Operator
===

Once the dev-script environment is up-and-running, and the [bmo][bmo] has been deployed, you can
proceed to deploying the Assisted Installer Operator following [this documentation](./operator).
At the end of this process you will have the assisted-service running under the `assisted-installer`
namespace.

Take notice of the [assisted-service-patch-image][aspi-custom] customization if you are testing a local
assisted-service image. More on this in the operator docs. Another, more hacky option, is to
manually modify the image for the `assisted-service` deployment after it's been deployed. Favor the
customization way over this.


```
[dev@edge-10 dev-scripts]$ oc get pods
NAME                                READY   STATUS    RESTARTS   AGE
assisted-service-7d45ccf4f8-88pw4   1/1     Running   0          18h
ocp-metal-ui-7967c757b4-fdxph       1/1     Running   0          21h
postgres-6bd4dc768f-54pgz           1/1     Running   0          21h
```

Tailing the logs for the `assisted-service` pod should not show any errors.

Creating BareMetalHost resources
===

The [baremetal operator][bmo] creates the `BareMetalHost` resources for the existing nodes
automatically. For scenarios using extra worker nodes (like SNO), it will be necessary to create
`BareMetalHost` resources manually. Luckily enough, `dev-scripts` is one step ahead and it has
prepared the manifest for us already.

```
less ocp/ostest/extra_host_manifests.yaml
```

You can modify this manifest to disable inspection, and cleaning. Here's an example on what it would look like:

```
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: ostest-worker-0
  namespace: assisted-installer
  annotations:
    # BMAC will add this annotation if not present
    inspect.metal3.io: disabled
  labels:
    infraenvs.agent-install.openshift.io: "bmac-test"
spec:
  online: true
  bootMACAddress: 00:ec:ee:f8:5a:ba
  # Disable automated cleaning mode
  automatedCleaningMode: disabled
  bmc:
    address: ....
    credentialsName: bmc-secret
```

The above is an optional step. If skipped, the `BareMetalHost` will boot IPA and spend some time in
the inspecting phase when the manifest is applied. There's an [open PR](https://github.com/openshift-metal3/dev-scripts/pull/1193/)
that will allow setting an environment variable to have inspection disabled by `dev-scripts` by default.


[bmo]: https://github.com/openshift/baremetal-operator
[local-bmo]: https://github.com/openshift-metal3/dev-scripts/blob/master/metal3-dev/local-bmo.sh
[aspi-custom]: https://github.com/openshift/assisted-service/blob/master/config/default/assisted-service-patch-image.yaml


Creating ClusterDeployment and InfraEnv
==

Before deploying the ClusterDeployment, make sure you have created a secret with your pull-secret.
In this environment it's called `my-pull-secret`

```
kubectl create secret -n assisted-installer generic my-pull-secret --from-file=.dockerconfigjson=pull_secret.json
```

At this point, it's possible to deploy the SNO cluster by modifying and applying [the SNO manifest](./crds/clusterDeployment-SNO.yaml)
