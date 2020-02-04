---
name: Feature request
about: Suggest an idea for this project
title: ''
labels: 'enhancement'
assignees: ''

---

<!-- Please answer these questions before submitting your feature request. Thanks! -->

**Is your feature request related to a problem? Please describe.**
<!-- A clear and concise description of what the problem is. Ex. I'm always frustrated when [...] -->

**Describe the solution you'd like**
<!-- A clear and concise description of what you want to happen. -->

**Describe alternatives you've considered**
<!-- A clear and concise description of any alternative solutions or features you've considered. -->

**What version of eunomia are you using?**

<details><summary><code>kubectl exec $EUNOMIA_POD curl localhost:8383/metrics</code> Output</summary><br><pre>
$ kubectl get -n eunomia-operator endpoints/eunomia-operator -o jsonpath='{.subsets[*].addresses[*].targetRef.name}' | xargs -I% kubectl exec -n eunomia-operator % -- curl -sS localhost:8383/metrics | grep eunomia_build_info

</pre></details>

eunomia version:

**Additional context**
<!-- Add any other context or screenshots about the feature request here. -->
