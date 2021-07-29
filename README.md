# nanovms
Test of nanovms with Golang

This code is based on some that I wrote for an evaluation and for a personal project.

The goal of this small project is to explore the use of the ops and nanovms
unikernel to deploy as an instance image on a cloud platform. I am using GCP in
this case. Although it may run as compiled it is not ready to deploy yet.

I would like to explore both deployment as a standalone image and as an image in
a managed instance cluster.

See https://nanovms.com

## Sample run

```
ops run -p 8000 ./nanoapplinux
 100% |████████████████████████████████████████|  [0s:0s]
 100% |████████████████████████████████████████|  [0s:0s]
booting /Users/ian/.ops/images/nanoapplinux.img ...
en1: assigned 10.0.2.15
Serving transactions on port 8000en1: assigned FE80::D809:D5FF:FE5A:637B
```