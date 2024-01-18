# gitops-operator test statistics
## FLAKY TESTS: Failed test scenarios in past 14 days
| Failure Score<sup>*</sup> | Failures | Test Name | Last Seen | PR List and Logs 
|---|---|---|---|---|
| 120 | 4 | 1-073_validate_rhsso  |  | 3: [#643](https://github.com/redhat-developer/gitops-operator/pull//643)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/643/pull-ci-redhat-developer-gitops-operator-master-v4.14-kuttl-parallel/1747569040978612224/build-log.txt), [2](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/643/pull-ci-redhat-developer-gitops-operator-master-v4.14-kuttl-parallel/1747604380435091456/build-log.txt)</sup> [#641](https://github.com/redhat-developer/gitops-operator/pull//641)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/641/pull-ci-redhat-developer-gitops-operator-master-v4.14-kuttl-parallel/1747132563488509952/build-log.txt)</sup> [#631](https://github.com/redhat-developer/gitops-operator/pull//631)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/631/pull-ci-redhat-developer-gitops-operator-master-v4.14-kuttl-parallel/1742908912308326400/build-log.txt)</sup> 
| 40 | 2 | 1-035_validate_argocd_secret_repopulate  |  | 2: [#641](https://github.com/redhat-developer/gitops-operator/pull//641)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/641/pull-ci-redhat-developer-gitops-operator-master-v4.13-kuttl-sequential/1747132560128872448/build-log.txt)</sup> [#631](https://github.com/redhat-developer/gitops-operator/pull//631)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/redhat-developer_gitops-operator/631/pull-ci-redhat-developer-gitops-operator-master-v4.12-kuttl-sequential/1742908908273405952/build-log.txt)</sup> 



<sup>*</sup> - Failure score is an arbitrary severity estimate, and is approximately `(# of PRs the test failure was seen in * # of test failures) / (days since failure)`. See code for full algorithm -- PRs welcome for algorithm improvements.

