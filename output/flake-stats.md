# gitops-operator test statistics
## FLAKY TESTS: Failed test scenarios in past 14 days
| Failure Score<sup>*</sup> | Failures | Test Name | Last Seen | PR List and Logs 
|---|---|---|---|---|
| 80 | 4 | 1-040_validate_quoted_RBAC_group_names  |  | 2: [#545](https://github.com/redhat-developer/gitops-operator/pull//545)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/545/pull-ci-redhat-developer-gitops-operator-master-v4.10-kuttl-sequential/1684519612483375104/build-log.txt)</sup> [#530](https://github.com/redhat-developer/gitops-operator/pull//530)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/530/pull-ci-redhat-developer-gitops-operator-master-v4.10-kuttl-sequential/1684179541662109696/build-log.txt), [2](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/530/pull-ci-redhat-developer-gitops-operator-master-v4.11-kuttl-sequential/1684179550918938624/build-log.txt)</sup> 
| 40 | 2 | 1-078_validate_default_argocd_consoleLink  |  | 2: [#573](https://github.com/redhat-developer/gitops-operator/pull//573)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/573/pull-ci-redhat-developer-gitops-operator-master-v4.10-kuttl-sequential/1681900925876178944/build-log.txt)</sup> [#559](https://github.com/redhat-developer/gitops-operator/pull//559)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/559/pull-ci-redhat-developer-gitops-operator-master-v4.11-kuttl-sequential/1682401830471995392/build-log.txt)</sup> 



<sup>*</sup> - Failure score is an arbitrary severity estimate, and is approximately `(# of PRs the test failure was seen in * # of test failures) / (days since failure)`. See code for full algorithm -- PRs welcome for algorithm improvements.

