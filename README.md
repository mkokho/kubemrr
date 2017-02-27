# kubemrr

`kubemrr` mirrors description of Kubernetes resources for blazingly fast auto-completion.

Feel free to suggest improvements, I'm here to work on them 👍

# Details

Bash  has the _programmable completion_ feature which permits typing a partial command, 
then pressing the [TAB] key to auto-complete the command sequence.To use the feature, you need to write a completion script 
and place it in /etc/bash_completion.d folder ([docs](http://www.tldp.org/LDP/abs/html/tabexpansion.html),
[tutorial](https://debian-administration.org/article/316/An_introduction_to_bash_completion_part_1)). 
Fortunately, `kubectl` comes with a command to generate completion script ([kubectl completion](http://kubernetes.io/docs/user-guide/kubectl/kubectl_completion/)). 
The script works extremely well, but slow, because each time you hit [TAB] it sends a request 
to Kubernetes API Server. If your server is on a different continent you might wait for up to 2 seconds. 
To reduce the delay, `kubemrr` keeps names of resources locally. See how to make a completion script which talks to `kubemrr` 
instead of real Kubernetes API server in the example below.

![b](https://cloud.githubusercontent.com/assets/10990119/20454663/0666ec22-aeac-11e6-9d7d-550313fb4b60.gif)

# Example

To start watching servers, give context names from your kubeconfig file:
```
kubemrr watch dev prod
```

To make completion script that talks to `kubemrr` shell:
```
alias kus='kubectl --context us'
kubemrr completion bash --kubectl-alias=kus > kus
sudo cp kus /etc/bash_completion.d
```

Note that you need to have bash completion installed. It shoud be available on a Linux distribution. On a Mac, 
install with `brew install bash-completion`.

Replace `bash` with `zsh` in the above command to generate completion script for `zsh` shell.

To test it:
```
source kus
kus get po [TAB][TAB]
kus get svc [TAB][TAB]
kus get deployments [TAB][TAB]
kus get configmaps [TAB][TAB]
kus get namespaces [TAB][TAB]
kus get nodes [TAB][TAB]
```

To make completion script that talks to `kubemrr` that is running on different host (use IP to save time on name resolution):
```
kubemrr completion bash --address=10.5.1.6 --kubectl-alias=kus > kus
```

# Download
- OSX: 
```
curl -O https://raw.githubusercontent.com/mkokho/kubemrr/v1.2.1/releases/darwin/amd64/kubemrr
```

- Linux: 
```
curl -O https://raw.githubusercontent.com/mkokho/kubemrr/v1.2.1/releases/linux/amd64/kubemrr
```
