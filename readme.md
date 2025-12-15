# Repo for the IBM Storage Virtualize — IP Quorum application Service

This repository provides resources and guides, for running the IP Quorum application for IBM Storage Virtualize. It includes multiple deployment options and tips for managing the service effectively.

## 📘 About the IP Quorum Application for IBM Storage Virtualize

🏗 Architecture
Here’s a simplified view of how IP Quorum interacts with IBM Storage Virtualize clusters:

<img src="ipquorum-systemd/images/ipquorum-smal.png" alt="drawing" style="width:600px;"/>

-----

A quorum device is used to break a tie when a SAN fault occurs, when exactly half of the nodes that were previously a member of the cluster are present

The IP quorum application is a Java application that runs on a separate server or host. (This can be physical or Virtual Machine.)
An IP quorum application is used in IP networks to resolve failure scenarios where half the control canisters/nodes on the cluster become unavailable.

The application determines which nodes or enclosures can continue processing host operations and avoids a split cluster, where both halves of the system continue to process I/O independently.

IBM Storage Virtualize powers the IBM Storage FlashSystem and IBM SVC


## 🚀 Deployment Optionsfor IP Quorum Service 

* [Guides for Running IP Quorum as a Systemd on linux](ipquorum-systemd/readme-ipquorum-systemd.md)

* [Guides for Running IP Quorum as a Container on linux](ipquorum-container/ibm-virtualize-ipquorum-container.md)

* [Guides for Downloading the IPQuorum with Script trough RestAPI](ipquorum-download/ipquorum-download-script-readme.md)

* [Guides for Downloading the IPQuorum Manually trough RestAPI](ipquorum-download/ipquorum-download-readme.md)


## 💡 Tips and Triks



## Downloading the Java app from Storage Virtualize Box

### RestAPI - New from IBM Storage Virtualize Code 8.6.1
Starting with IBM Storage Virtualize Code 8.6.1, you can download the IP Quorum application using the REST API.
Check out the guide in the repo:* [Guides for Downloading the IPQuorum Manually trough RestAPI](ipquorum-download)
