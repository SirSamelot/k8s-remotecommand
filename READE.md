# Remote Command PoC

Proof of Concept for issuing remote commands to kubernetes pods

## Notes
* busybox doesn't use `bash`.  It uses `ash`.  So use the following to get an interactive shell:
```kubectl exec -it bb /bin/ash```
