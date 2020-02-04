---
name: Bug report
about: Create a bug report to help us improve
title: ''
labels: 'bug'
assignees: ''

---

<!-- Please answer these questions before submitting your bug report. Thanks! -->

**What version of eunomia are you using?**

<details><summary><code>kubectl exec $EUNOMIA_POD curl localhost:8383/metrics</code> Output</summary><br><pre>
$ kubectl get -n eunomia-operator endpoints/eunomia-operator -o jsonpath='{.subsets[*].addresses[*].targetRef.name}' | xargs -I% kubectl exec -n eunomia-operator % -- curl -sS localhost:8383/metrics | grep eunomia_build_info

</pre></details>

eunomia version:

**Does this issue reproduce with the latest release?**



**What operating system and processor architecture are you using (`kubectl version`)?**

<details><summary><code>kubectl version</code> Output</summary><br><pre>
$ kubectl version

</pre></details>

**What did you do?**

<!--
If possible, provide a recipe for reproducing the error.
A detailed sequence of steps describing what to do to observe the issue is good.
A complete runnable bash shell script is best.
-->



**What did you expect to see?**



**What did you see instead?**
