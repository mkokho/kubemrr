# kubemrr

`kubemrr` mirrors description of Kubernetes resources for blazingly fast auto-completion.

`kubectl` comes with a command to generate completion script for `bash` shell. 
The script works extremely well, but slow, because each time you hit [TAB] it sends a request 
to Kubernetes API Server. If your server is on a different continent you might wait for up to 2 seconds.
`kubemrr` keeps a description of resources locally. We will make completion script talk to `kubemrr` 
instead of real Kubernetes API server.

# Example

To start watching a server:

- `kubemrr -p 34000 watch https://kube-us.example.org`

To get names of the pods, services or deployments that are active on the server:

- `kubemrr -p 34000 get pods`
- `kubemrr -p 34000 get services`
- `kubemrr -p 34000 get deployments`

# Enabling auto-completion

Most shells allow to program command completion. All you need to do is supply a completion script. 
Below is a guide for `bash` shell

## Bash

- `cd tmp`
- `kubectl completion bash > ./kubectl_completion`
- `sed -i s/'kubectl get $(__kubectl_namespace_flag) -o template --template="${template}"'/'kubemrr get'/g kubectl_completion` 
- `sudo mv kube_completion /etc/bash_completion.d/kubectl`
- `source /etc/bash_completion.d/kubectl`

To test it, start watching a server, and then type `kubectl get po [TAB][TAB]`

## Bash with alias

There is few more place to change in `/etc/bash_completion.d/kubectl` file. 
Let's say your alias is `kus` and you keep mirror on port 33003:

- `kubectl completion bash > ./kus`
- `sed -i s/'kubectl get $(__kubectl_namespace_flag) -o template --template="${template}"'/'kubemrr -p $kubemrr_port get'/g kus`
- `sed -i s/'local commands=("kubectl")'/'local commands=("kus")'/g kus`
- `sed -i s/'_kubectl()'/'_kus()'/g kus`
- `sed -i s/'local c=0'/'local c=0\n    local kubemrr_port=33003'/g kus`
- `sed -i s/'__start_kubectl kubectl'/'__start_kus kus'/g kus`

It is possible to create script for multiple aliases. However, it is easy to break things because 
function name are mostly the same. Nevertheless, you can experiment. Just run the above commands on 
different files, and call `source` on them. Then test it. If you broke things, re-open terminal, 
and retry ;)

