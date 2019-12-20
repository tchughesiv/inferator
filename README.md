# Inferator

[![Go Report](https://goreportcard.com/badge/github.com/tchughesiv/inferator)](https://goreportcard.com/report/github.com/tchughesiv/inferator)

## Requirements

- go 1.13.x
- operator-sdk 0.12.0

## Build

```bash
make
```

## Upload to a container registry

e.g.

```bash
docker push quay.io/tchughesiv/inferator:<version>
```

## Deploy to OpenShift 4 using OLM

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Push the operator bundle to your quay application repository as follows:

```bash
REGISTRY_NS=<registry namespace in quay.io>
operator-courier push deploy/catalog_resources/community ${REGISTRY_NS} inferator-operator $(go run getversion.go -operator) "basic XXXXXXXXX"
```

Note that the push command does not overwrite an existing repository, and it needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

```bash
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: inferator-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: ${REGISTRY_NS}
  displayName: "Inferator Operators"
  publisher: "Red Hat"
EOF
```

## Development

Change log level at runtime w/ the `DEBUG` environment variable. e.g. -

```bash
make vet
DEBUG="true" operator-sdk up local --namespace=<namespace>
```

Also at runtime, change registry for rhpam ImageStreamTags -

```bash
INSECURE=true REGISTRY=<registry url> operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to generate, vet, format, and test your code. This all can be done with one command.

```bash
make test
```

## Test zenithr

```bash
curl -v http://example-operationrule-tommy4.apps-crc.testing/ \
 -H "Content-Type: application/json" \
 --data-binary "@test/deployment-replicas1.json" | jq

curl -v http://example-operationrule-tommy4.apps-crc.testing/ \
 -H "Content-Type: application/json" \
 --data-binary "@test/deployment-replicas3.json" | jq
```
