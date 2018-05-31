# Remote Command PoC

Proof of Concept for issuing remote commands to kubernetes pods

## Instructions

1. Start the busybox pod

    ```kubectl create -f busybox.yml```

2. Run the PoC script

    ```go run rc.go```

## Notes

* busybox doesn't use `bash`.  It uses `ash`.  So use the following to get an interactive shell:

    ```kubectl exec -it bb /bin/ash```
