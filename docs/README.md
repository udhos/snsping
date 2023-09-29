# Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

    helm repo add snsping https://udhos.github.io/snsping

Update files from repo:

    helm repo update

Search snsping:

    helm search repo snsping -l --version ">=0.0.0"
    NAME            CHART VERSION	APP VERSION	DESCRIPTION
    snsping/snsping	0.1.0        	0.0.0      	Install snsping helm chart into kubernetes.

To install the charts:

    helm install my-snsping snsping/snsping
    #            ^          ^       ^
    #            |          |        \_______ chart
    #            |          |
    #            |           \_______________ repo
    #            |
    #             \__________________________ release (chart instance installed in cluster)

To uninstall the charts:

    helm uninstall my-snsping

# Source

<https://github.com/udhos/snsping>
