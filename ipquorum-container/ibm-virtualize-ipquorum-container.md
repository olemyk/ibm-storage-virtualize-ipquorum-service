# How to configure IBM Storage Virtualize — IP-Quorum app to run inside a container. 🚢



Let's run the IP-Quorum/java app inside a container using PODMAN or Docker.

## Some information about the IP-Quorum application for Spectrum Virtualize.

<img src="../ipquorum-systemd/images/ipquorum-smal.png" alt="drawing" style="width:600px;"/>

* A quorum device is used to break a tie when a SAN fault occurs.
When exactly half of the nodes that were previously a member of the cluster are present

* The IP quorum application is a Java application that runs on a separate server or host. (This can be a physical or Virtual Machine and now a container.)

* An IP quorum application is used in IP networks to resolve failure scenarios where half the control canisters/nodes on the cluster become unavailable.
The application determines which nodes or enclosures can continue processing host operations and avoids a split cluster, where both halves of the system continue to process I/O independently.


## There is also two different option when creating the IP-Quorum Java application in Storage Virtualize.

These have different requirements, these are listed below.

    Tie-break (no metadata)

    Firewall — Port 1260 /TCP
    Round trip latency of 80ms
    Network bandwidth of 2MB/s

    A new app is required, if cluster size changes e.g., node added/removed
    Security of the host running the app — authorized access only!
    Interop Matrix of supported OSs and Java Variants
    Max 5 apps
    
    Cluster Recovery (metadata)
    Everything from tie-break, also:
    Increased requirement for network bandwidth to 64MB/s
    250MB of disk space
    Only one app per IP address.


## What do you need ✅

Make sure you have these before starting:

    Host/VM to serve the IP-Quorum container
    Podman installed.
    IP connection to the Spectrum Virtualize cluster (SpecV).
    - Port 1260/TCP from the IP-Quorum host/container.
    If you want to fetch the IP-Quorum application with SCP, you need an SCP connection from your IP-Quorum VM to the Spectrum Virtualize cluster, if not download the ip_quorum.jar file manually.
    For more information about IP-Quorum Application and config check KC
    The image repo for the container is from RedHat or Docker.

## Create the IP-Quorum Container 🗳

! This is a work in progress, I'm looking into how we should do the user access and what container image is the best one. so open for ideas!

   Create a user and group to run the service and Container as (This is Optional if you don’t want to run with root.)

    Note: When not defining the UID and GUI in the command below the UID/GUI will be generated.

1. Create the group ipquorum

        getent group ipquorum >/dev/null || groupadd -r ipquorum
        
2. Creates the User.
        
        getent passwd ipquorum >/dev/null || useradd -c "IP-Quorum service account" -g ipquorum -s /sbin/nologin -r -d / ipquorum

        This adds user to group ipquorum,
        shell is nologin, -r = system account, home-dir = ipquorum.
    

3. Check that the user is created, either with ID or passwd file.

        id ipquorum
        cat /etc/passwd | grep ipquorum
        write down the ID and GID
    
4. Create a directory that will be mapped into the container.

    `mkdir -p /srv/ipquorum/logs`

5.  Change the owner of the directory to the user we created,
    In this case ipquorum.

    `chown ipquorum /srv/ipquorum/logs`


6. You can Copy in the ip_quorum.jar file from the Spectrum Virtualize box with SCP.
    Users on SpecV with Monitor role can also copy over the file.

    `scp ipquorumscp@10.33.7.56:/dumps/ip_quorum.jar /srv/ipquorum/logs`
    


7. Create the ip_quorum startup script, copy the whole section to the terminal.

        cat <<'EOF' > /srv/ipquorum/start.sh
        #! /bin/sh -
        #
        #Start service:
        cd /srv/ipquorum/logs
        /usr/bin/java -jar /srv/ipquorum/logs/ip_quorum.jar &
        #
        # Send all logs to stdout, so that we get them in the journal:
        sleep 5
        tail -n +1 -F /srv/ipquorum/logs/ip_quorum.log.0
        EOF

8. Change the rights for the startup script, so the user can start it.
    `chmod +x /srv/ipquorum/start.sh`

9. Let’s start the container

    Have created the following example:

        podman run -d --name specv-ipquorum --log-driver=journald --user=989:985 --hostname=specv-ipquorum-container -w /srv/ipquorum/logs --volume /srv/ipquorum:/srv/ipquorum docker.io/library/openjdk:14 /srv/ipquorum/start.sh

    --

        An explanation for the options:
        -d run container in the background
        --name name of container
        --log-driver= log to journalctl when the systemd services is started
        --user= run with the created users ID and group. (Maybe overkill)
        --hostname= of the container, so it will show up with the name in SpecV GUI
        -w working directory, Is not needed when then script CD into the /srv/ipquorum/logs
        --volume to map into the container, in this case, it will have access to the startup script, java app file and directory to save logs, config and metadata from SpecV.
        openjdk:14 is from docker hub, simple OpenJDK image, this is built with several versions, one thing we need to check is what version we are using and what is supported. Version 8.4 of SpecV doc it saying OpenJDK 14 is supported.

    `/srv/ipquorum/start.sh` is the startup script that we created earlier to start the ipquorum app and then send logs out to the journal


    
    Optional: Example with OpenJDK image from RedHat, this contains also JBoss image

        podman run -d --name specv-ipquorum2 --log-driver=journald --user=989:985 --hostname=specv-ipquorum-container -w /srv/ipquorum/logs --volume /srv/ipquorum:/srv/ipquorum registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift:latest /srv/ipquorum/start.sh

10. Create a systemd service so the container starts and stops the container with the system.

        podman generate systemd --new --name specv-ipquorum > /etc/systemd/system/specv-ipquorum.service

    Then restart the service:
        `systemctl restart specv-ipquorum`

11. Check Status of the Systemd

        systemctl status specv-ipquorum

    As we can see, the service is running. and you should have details about Command=HEARTBEAT_RESPONSE.

        [root@podman01 ~]# systemctl status specv-ipquorum
        ● specv-ipquorum.service - Podman container-specv-ipquorum.service
        Loaded: loaded (/etc/systemd/system/specv-ipquorum.service; disabled; vendor preset: disabled)
        Active: active (running) since Mon 2022-03-07 20:06:37 CET; 1min 13s ago
            Docs: man:podman-generate-systemd(1)
        Process: 575382 ExecStopPost=/usr/bin/podman rm -f --ignore --cidfile=/run/specv-ipquorum.service.ctr-id (code=exited, status=0/SUCCESS)
        Process: 575218 ExecStop=/usr/bin/podman stop --ignore --cidfile=/run/specv-ipquorum.service.ctr-id (code=exited, status=0/SUCCESS)
        Process: 575426 ExecStartPre=/bin/rm -f /run/specv-ipquorum.service.ctr-id (code=exited, status=0/SUCCESS)
        Main PID: 575545 (conmon)
            Tasks: 2 (limit: 49461)
        Memory: 1.6M
        CGroup: /system.slice/specv-ipquorum.service
                └─575545 /usr/bin/conmon --api-version 1 -c 4758337b527dcd94e14461bc427175a98b57af7f9d6d96446f1ca5248a521462 -u 4758337b527dcd94e14461bc427175a98b57af7f9d6d96446f1ca5248a521462 -r /usr/bin/runc -b /var/lib/containers/storage/o>Mar 07 20:07:31 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:07:31:118 10.33.7.59 [18] FINE: <Msg [protocol=1, sequence=5, command=HEARTBEAT_REQUEST, length=0]
        Mar 07 20:07:31 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:07:31:121 10.33.7.59 [18] FINE: >Msg [protocol=1, sequence=5, command=HEARTBEAT_RESPONSE, length=0]
        Mar 07 20:07:46 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:07:46:081 10.33.7.57 [15] FINE: <Msg [protocol=1, sequence=6, command=HEARTBEAT_REQUEST, length=0]
        Mar 07 20:07:46 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:07:46:082 10.33.7.58 [16] FINE: <Msg [protocol=1, sequence=7, command=HEARTBEAT_REQUEST, length=0]

    --

    We can now see that Java App inside the container has a connection from the logs...
    We want also just to be sure, to check the status in the SpecV GUI. The IP-Address is container network-magic, and hostname is the one we set with - -hostname=

## Information and Troubleshooting

        
 To check the Logs, you can run the journalctl command with the systemd name.

`journalctl -u specv-ipquorum -r`


Example:

    [root@podman01 ~]# journalctl -u specv-ipquorum -r
    -- Logs begin at Thu 2022-03-03 23:22:19 CET, end at Mon 2022-03-07 20:10:31 CET. --
    Mar 07 20:10:31 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:10:31:085 10.33.7.60 [17] FINE: >Msg [protocol=1, sequence=17, command=HEARTBEAT_RESPONSE, length=0]
    Mar 07 20:10:31 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:10:31:084 10.33.7.57 [15] FINE: >Msg [protocol=1, sequence=17, command=HEARTBEAT_RESPONSE, length=0]
    Mar 07 20:10:31 podman01.oslo.forum.ibm.com specv-ipquorum[575545]: 2022-03-07 19:10:31:084 10.33.7.58 [16] FINE: >Msg [protocol=1, sequence=18, command=HEARTBEAT_RESPONSE, length=0]

You can also check the logs directly in the container.

    [root@podman01 ~]# podman logs specv-ipquorum
    === IP quorum ===
    Name set to null.
    Successfully parsed the configuration, found 4 nodes.
    Trying to open socket
    Trying to open socket
    Trying to open socket
    Trying to open socket
    Handshaking
    Handshaking
    Handshaking
    Handshaking
    Creating UID
    Waiting for UID
    Waiting for UID
    Waiting for UID
    *Connecting
    *Connecting
    *Connecting
    Connected to 10.33.7.60
    Connected to 10.33.7.57
    Connected to 10.33.7.59
    Connected to 10.33.7.58
    2022–03–07 19:59:31:605 Quorum CONFIG: === IP quorum ===
    2022–03–07 19:59:31:607 Quorum CONFIG: Name set to null.
    2022–03–07 19:59:32:076 Quorum FINE: Node 10.33.7.57:1,260 (0, 0)
    2022–03–07 19:59:32:086 Quorum FINE: Node 10.33.7.58:1,260 (0, 0)
    2022–03–07 19:59:32:087 Quorum FINE: Node 10.33.7.60:1,260 (0, 0)
    2022–03–07 19:59:32:088 Quorum FINE: Node 10.33.7.59:1,260 (0, 0)
    2022–03–07 19:59:32:089 Quorum CONFIG: Successfully parsed the configuration, found 4 nodes.
    2022–03–07 19:59:32:094 10.33.7.59 [17] INFO: Trying to open socket
    2022–03–07 19:59:32:092 10.33.7.57 [14] INFO: Trying to open socket
    2022–03–07 19:59:32:092 10.33.7.58 [15] INFO: Trying to open socket
    2022–03–07 19:59:32:093 10.33.7.60 [16] INFO: Trying to open socket
    2022–03–07 19:59:32:185 10.33.7.60 [16] INFO: Handshaking
    2022–03–07 19:59:32:185 10.33.7.57 [14] INFO: Handshaking
    2022–03–07 19:59:32:185 10.33.7.59 [17] INFO: Handshaking
    2022–03–07 19:59:32:185 10.33.7.58 [15] INFO: Handshaking
    2022–03–07 19:59:32:195 10.33.7.58 [15] FINE: <Msg [protocol=3, sequence=0, command=HANDSHAKE_REQUEST, length=1] Data [features_bitmap=0]
    2022–03–07 19:59:32:195 10.33.7.59 [17] FINE: <Msg [protocol=3, sequence=0, command=HANDSHAKE_REQUEST, length=1] Data [features_bitmap=0]
    2022–03–07 19:59:32:195 10.33.7.57 [14] FINE: <Msg [protocol=3, sequence=0, command=HANDSHAKE_REQUEST, length=1] Data [features_bitma

If you want to log into and check the container content, you can use the exec — it with bash.
    Example: how to check the java version and directory

        [root@podman01 ~]# podman exec -it specv-ipquorum bashbash-4.4$ java -version
        openjdk version “14.0.2” 2020–07–14
        OpenJDK Runtime Environment (build 14.0.2+12–46)
        OpenJDK 64-Bit Server VM (build 14.0.2+12–46, mixed mode, sharing)bash-4.4$ ls
        ip_quorum.jar ip_quorum.log.0 ip_quorum.log.0.lck
        bash-4.4$ pwd
        /srv/ipquorum/logs

Contribution: ❗ 🤔
As I say, this is a work in progress, And I’m still looking into how we should do the user access better and what container image is the best one. so please reach out if you have ideas.