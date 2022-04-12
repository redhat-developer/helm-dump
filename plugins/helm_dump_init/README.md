# helm_dump_init

This is a `crane-lib` plugin meant to be used during the initial `helm dump init` process.

To test in the command line the reader should first install `yq` to convert the YAML test files into JSON (at plugin 
runtime this process is handled by `crane-lib` APIs); this can be done by executing  the following command: 
`pip3 install yq`. The example below also assume `jq` is installed in the system.

```text
$ cat test/test_run/nginx-deployment-without-labels.yaml | yq | ./helm_dump_init | jq
{
  "version": "v1",
  "patches": [
    {
      "op": "replace",
      "path": "/metadata/name",
      "value": "nginx-deployment-{{ .Release.Name }}"
    },
    {
      "op": "add",
      "path": "/metadata/labels",
      "value": {}
    },
    {
      "op": "replace",
      "path": "/metadata/labels/app.kubernetes.io~1name",
      "value": "{{ template \"fullname\" $ }}"
    },
    {
      "op": "replace",
      "path": "/metadata/labels/app.kubernetes.io~1instance",
      "value": "{{ $.Release.Name }}"
    }
  ]
}

```

## References

- http://jsonpatch.com/
- https://www.rfc-editor.org/rfc/rfc6901#section-3
- https://github.com/konveyor/gitops-primer/tree/main/export/plugins`