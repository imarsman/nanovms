# nanovms
Test of nanovms with Golang. The goal of this project is to understand how to
use ops and the gcloud CLI to build images, get them to the cloud, build
instances, and successfully run them with proper port settings. Once this works
it should be possible to create a cluster using an image and set the cluster to
scale up and down with load.

This code is based on some that I wrote for an evaluation and for a personal project.

The goal of this small project is to explore the use of the ops and nanovms
unikernel to deploy as an instance image on a cloud platform. I am using GCP in
this case. Although it may run as compiled it is not ready to deploy yet.

I would like to explore both deployment as a standalone image and as an image in
a managed instance cluster.

See https://nanovms.com

You can see things as they stand [here](http://35.211.243.30:8000). This may
intermittently be unavailable if I am doing work on the code and redeploying.

## What works

- building native and linux
- creating image the first time
  - image on GCP not deleted properly first so that does not get updated
- creating instance from image
- accessing unikernal instance via http://theip/transactions

## What does not work

- proper shutdown of existing instance and/or deletion
- deletion of GCP image prior to creation of a new one
- Having a new instance have opened ports without manual intervention

Once the proper steps are handled in the proper order it should work to automate
building, image creation, instance creation, etc.