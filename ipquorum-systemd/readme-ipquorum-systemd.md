# How to configure a systemd service for the IBM Spectrum Virtualize — IP-Quorum app.

I have already written a guide on how to use my Ansible role for installing and configuring IP-Quorum Service for Storage Virtualize. 
In this Guide I show how to this can be done manually on Centos/RedHat by creating a systemd service
IBM Storage Virtualize (San Volume Controller, SVC, Storwize, FlashSystem,)


# Some information about the IP-Quorum application for Spectrum Virtualize.

<img src="images/ipquorum-smal.png" alt="drawing" style="width:600px;"/>

-----

A quorum device is used to break a tie when a SAN fault occurs, when exactly half of the nodes that were previously a member of the cluster are present
The IP quorum application is a Java application that runs on a separate server or host. (This can be physical or Virtual Machine.)
An IP quorum application is used in IP networks to resolve failure scenarios where half the control canisters/nodes on the cluster become unavailable.
The application determines which nodes or enclosures can continue processing host operations and avoids a split cluster, where both halves of the system continue to process I/O independently.


## There is two different option when creating the IP-Quorum Java application.

These have different requirements anf functions, these are listed below.
Please check Storage Virtualize IBM Doc for updated requirements.

https://www.ibm.com/docs/en/flashsystem-7x00/9.1.0?topic=quorum-ip-application
https://www.ibm.com/support/pages/node/7013877

**Tie-break (nometadata)**

        Firewall — Port 1260 /TCP
        Round trip latency of 80ms
        Network bandwidth of 2MB/s
        New app, if cluster size changes e.g., node added/removed
        Security of the host running the app — authorized access only!
        Interop Matrix of supported OSs and Java Variants
        Max 5 apps

**Cluster Recovery (metadata)**

        Everything from tie-break, also:
        Increased requirement for network bandwidth to 64MB/s
        250MB of disk space
        Only one app per IP address.

## What do you need to run the IP-Quorum ✅

Make sure you have these before start:

    - Host/VM to serve the IP-Quorum service
    - IP connection to the Spectrum Virtualize cluster (SV).
        - Port 1260/TCP from the IP-Quorum host
    - SSH connection from your Ansible controller to the IP-Quorum host.
    - If you want to fetch the IP-Quorum application with SCP, you need SCP connection from your IPQuorum VM to the Storage Virtualize cluster, if not download the ip_quorum.jar file manually.

For more information about IP-Quorum Application and config check IBM Doc

## Lest start Configuring IP-Quorum 🔨

1. **Firewall Rules**

    If you have strict firewall rules, you need to add the port 1260,

        firewall-cmd --add-port=1260/tcp
        firewall-cmd --add-port=1260/tcp --permanent

2. **Selinux**

    If you are running with default selinux then services will run under unconfined_t


3. **Create a user and group to run the service as (This is Optional if you don't want to run with nobody.)**

    Note: When not defining the UID and GUI in the command below the UID/GUI will be generated.

    Create the group ipquorum

        getent group ipquorum >/dev/null || groupadd -r ipquorum

    Creates the User.
    Add user to group ipquorum, shell is nologin, -r = system account, home-dir = ipquorum.

        getent passwd ipquorum >/dev/null || useradd -c "IP-Quorum service account" -g ipquorum -s /sbin/nologin -r -d / ipquorum 

    Check that the user is created with ID or passwd file.

        id ipquorum 
        uid=988(ipquorum) gid=984(ipquorum) groups=984(ipquorum)cat /etc/passwd | grep ipquorum
        ipquorum:x:988:984:IP-Quorum service account:/:/sbin/nologin

4. **Create ip-quorum directory.**
    
        mkdir -p /opt/IBM/ip-quorum/

5. Transfer the IP quorum application from the system to a directory on the host that is to run the IP quorum application.

    **Create the ip-quorum.jar with following options.**
    **
    
    **SpecV GUI:** In the management GUI, select Settings > System > IP Quorum and download the version of the IP quorum Java application into /opt/IBM/ip-quorum/
        
    **SpecV CLI:** You can also use the command-line interface (CLI) to enter the mkquorumapp command to generate an IP quorum Java application. The application is stored in the dumps directory of the system with the file name ip_quorum.jar.

    ssh superuser@IP-to-Flashsystem
    Command to create a new app:

        mkquorumapp -nometadata
        
    **SCP** 
    Using SCP to copy the ip_quorum.jar from Spectrum Virtualize to ip-quorum host
        
        scp superuser@specv-ip:/dumps/ip_quorum.jar /opt/IBM/ip-quorum
    **RestAPI**
    * [Guides for Downloading the IPQuorum with Script trough RestAPI](ipquorum-download/ipquorum-download-script-readme.md)

6. **Install JAVA**

    The supported java version is listed here. Supported Java and OS
    easiest way is just to install openjdk with yum install. Example:
    dnf install java

7. **Verify Java**
    
    "/usr/bin/java --version",  and it´s the correct version,
    Change the version using alternatives --config java

8. **Create the /opt/IBM/ip-quorum/log directory, and give it to “nobody” or a user that you want.**

        This is where the service will be running, and putting it’s logs:
        mkdir /opt/IBM/ip-quorum/log
        chown nobody /opt/IBM/ip-quorum/log
        If you created ipquorum user
        chown ipquorum /opt/IBM/ip-quorum/log
        If you using nobody user
        chown nobody /opt/IBM/ip-quorum/log

9. **Create a systemd unit file for IP-Quorum**

        vi /etc/systemd/system/ipquorum.service
        
    - Check and change if needed:
    - Working dir, ExecStart and users/group, nobody

    ```       
    [Unit]
    Description=IBM IP Quorum
    Documentation=https://www.ibm.com/docs/en/flashsystem-7x00/9.1.0?topic=quorum-ip-application
    After=local-fs.target network-online.target

    [Service]
    Type=simple
    WorkingDirectory=/opt/IBM/ip-quorum/log
    ExecStart=/usr/bin/java -jar /opt/IBM/ip-quorum/ip_quorum.jar
    Restart=on-failure
    PrivateTmp=yes
    InaccessibleDirectories=/home /root /var /boot
    ReadOnlyDirectories=/etc /usr
    User=ipquorum
    Group=ipquorum

    [Install]
    WantedBy=multi-user.target
    ```

10. Tell systemd to pick up this new service file using the following command: systemctl daemon-reload

        Enable service with: systemctl enable ipquorum
        Start service with: systemctl start ipquorum
        Check status with:systemctl status ipquorum

    The systemd should say active and logs should say connected to *

        ● ipquorum.service - IBM IP Quorum
        Loaded: loaded (/etc/systemd/system/ipquorum.service; enabled; vendor preset: disabled)
        Active: active (running) since Wed 2021-01-27 14:06:05 CET; 13min ago
            Docs: https://www.ibm.com/support/knowledgecenter/ST3FR7_8.3.1/com.ibm.storwize.v7000.831.doc/svc_ipquorumconfig.html
        Main PID: 5766 (java)
            Tasks: 24 (limit: 23713)
        Memory: 45.0M
        CGroup: /system.slice/ipquorum.service
                └─5766 /usr/bin/java -jar /opt/IBM/ip-quorum/ip_quorum.jarJan 27 14:06:05 centos8 java[5766]: Name set to null.
        Jan 27 14:06:06 centos8 java[5766]: Successfully parsed the configuration, found 4 nodes.
        Jan 27 14:06:06 centos8 java[5766]: Trying to open socket
        Jan 27 14:06:06 centos8 java[5766]: Trying to open socket
        Jan 27 14:06:11 centos8 java[5766]: Handshaking
        Jan 27 14:06:11 centos8 java[5766]: Creating UID
        Jan 27 14:06:11 centos8 java[5766]: Waiting for UID
        Jan 27 14:06:11 centos8 java[5766]: *Connecting
        Jan 27 14:06:11 centos8 java[5766]: Connected to 10.X.7.57
        Jan 27 14:06:11 centos8 java[5766]: Connected to 10.X.7.58

    From the SpecV Gui, the state should be Online.

11. If you want you can reboot and verify that the service is running afterwards.

    Note: If you want to run more than one IP-service,
    just create a new ip-quorum folder and a new service with a different name.
    Troubleshooting: 🕵

    Check log files under /opt/IBM/ip-quorum/log
        - If it says found nodes but the connection is unsuccessful, it´s most likely a network problem.
        To check what user the process is started with, check it with the following command:

        ps -ef | grep ipquorum
        ipquorum    5766       1  0 14:06 ?        00:00:03 /usr/bin/java -jar /opt/IBM/ip-quorum/ip_quorum.jar

    Selinux could create some issues, so try to change this to enforcing or disabled.
    The best practice is that services don’t own the files, however, if there is a problem starting the service and reading the ip_quorum.jar file, check permission.
    
    Example of how the logs should look like:

            [root@centos8 log]# cat ip_quorum.log.0
            2021–01–27 14:06:05:995 Quorum CONFIG: === IP quorum ===
            2021–01–27 14:06:05:997 Quorum CONFIG: Name set to null.
            2021–01–27 14:06:06:438 Quorum FINE: Node 10.xx.xx.57:1,260 (0, 0)
            2021–01–27 14:06:06:450 Quorum FINE: Node 10.xx.xx.58:1,260 (0, 0)
            2021–01–27 14:06:06:451 Quorum FINE: Node 10.xx.xx.60:1,260 (0, 0)
            2021–01–27 14:06:06:451 Quorum FINE: Node 10.xx.xx.59:1,260 (0, 0)
            2021–01–27 14:06:06:452 Quorum CONFIG: Successfully parsed the configuration, found 4 nodes.
            2021–01–27 14:06:06:456 10.x.7.57 [11] INFO: Trying to open socket
            2021–01–27 14:06:06:456 10.x.7.60 [13] INFO: Trying to open socket
            2021–01–27 14:06:06:457 10.x.7.58 [12] INFO: Trying to open socket
            2021–01–27 14:06:06:456 10.x.7.59 [14] INFO: Trying to open socket
            2021–01–27 14:06:11:515 10.x.7.58 [12] INFO: Handshaking
            2021–01–27 14:06:11:515 10.x.7.57 [11] INFO: Handshaking
            2021–01–27 14:06:11:533 10.x.7.58 [12] FINE: <Msg [protocol=3, sequence=0, command=HANDSHAKE_REQUEST, length=1] Data [features_bitmap=0]
            2021–01–27 14:06:11:535 10.x.7.57 [11] FINE: <Msg [protocol=3,

    The IPquorum seems to use a Random port (ephemeral range) localy and then connect up the SpecV with 1260


        [root@centos8 ~]# sudo lsof -i:1260
        COMMAND  PID     USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
        java    5766 ipquorum   11u  IPv6  38517      0t0  TCP centos8:39260->10.x.7.57:ibm-ssd (ESTABLISHED)
        java    5766 ipquorum   14u  IPv6  38520      0t0  TCP centos8:53272->10.x.7.58:ibm-ssd (ESTABLISHED)
        ------------------ 
        Check the connection with netstat and grep for java.[root@centos8 ~]# netstat -putan | grep java
        tcp6       0      0 10.x.3.230:53272       10.x.7.58:1260         ESTABLISHED 5766/java
        tcp6       0      0 10.x.3.230:39260       10.x.7.57:1260         ESTABLISHED 5766/java

    Just to check if we have Port 1260 open to the Service IP

        [root@centos8 ~]# curl -v telnet://10.x.7.57:1260
        * Rebuilt URL to: telnet://10.x.7.57:1260/
        *   Trying 10.x.7.57...
        * TCP_NODELAY set
        * Connected to 10.x.7.57 (10.x.7.57) port 1260 (#0)

    From the SpecV box. Use the command lsquorum to check if the connection is up, this can also be done from GUI.


        IBM_2145:SVC02:superuser>lsquorum
        quorum_index status id name          controller_id controller_name active object_type override site_id site_name
        0            online 13 fs840-1-lun_5 8             FS840-01        no     mdisk       no       1       site1
        1            online 3  fs840-2-lun_3 7             FS840-02        no     mdisk       no       2       site2
        2            online 4  fs840-2-lun_4 7             FS840-02        no     mdisk       no       2       site2
        3            online                                                yes    device      no               centos8/10.X.3.230

    If you want to check the network with Ping from SpecV to Ip-quorum host you can use Ping.
    ping -srcip4 “ServiceIP” “ IPQuorumIP”

        IBM_2145:SVC02:superuser> ping -srcip4 10.X.7.58 10.X.3.230
        PING 10.X.3.230 (10.X.3.230) from 10.X.7.58 : 56(84) bytes of data.



**Create a banner with IP-Quorum service information.**

Create a banner before login in through ssh

1. edit /etc/issue.net file and add banner text and save:

        vim /etc/issue.net

        ====================================================================
        IBM Spectrum Virtualize IP Quorum service
        ====================================================================

2. edit the sshd_config file, find the line beginning with Banner, edit it and save.

        vim /etc/ssh/sshd_config
        banner /etc/issue.net

3. Restart sshd

        systemctl restart sshd
        Adding a banner after login

    Edit and add the following text into file: vim /etc/motd

        ====================================================================
        The IP Quorom java app background service: ipquorum.service

            stop it with "systemctl stop ipquorum"
            start it with "systemctl start ipquorum"
            check it with "systemctl status ipquorum"

        The ip_quorum.jar file is located in /opt/IBM/ip-quorum/

        If needed to update. Stop the service, replace the ip_quorum.jar file
        and start the service.
        ===================================================================