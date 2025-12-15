
## 1.Download manually with curl trough RestAPI manually

Examples are divided into environment variable or manually changing the variables.

### First authenticate:


# Required variables (edit values)
```shell
export VIP="10.33.7.80"
export VUSERNAME="superuser"
export VPASSWORD="password"
export PARTNERSYSTEM="svc_cluster02"
```

1. Using the environment variables
    ```shell
    AUTH_RESPONSE=$(curl -ks -X POST "https://${IP}:7443/rest/v1/auth" \
      -H "accept: application/json" \
      -H "X-Auth-Username: ${VUSERNAME}" \
      -H "X-Auth-Password: ${VPASSWORD}" \
      -d "")
    ```
    2. Inspect the raw response (optional)
    `echo "${AUTH_RESPONSE}"`

    3. Extract token using jq (ensure jq is installed) 

      ```Shell
      TOKEN=$(echo "${AUTH_RESPONSE}" | jq -r '.token // .X_Auth_Token // .authToken // empty')
      ```

Manually: Change the IP and the Token you recived in step 1.
```shell
curl -ks -X POST "https://10.33.7.80:7443/rest/v1/auth" -H  "accept: application/json" -H  "X-Auth-Username: username" -H  "X-Auth-Password: password" -d ""
```



## 2. Create a new IPQuorum jar file.
IPQuorum is dumped into /dump folder of IBM Storage Virtualized config node. 

This calls mkquorumapp and sets your partnersystem. It assumes IPv4 only (ip_6: false, partnerip6: false) and that metadata is enabled (nometadata: false).

```shell
curl -vks -X POST "https://${IP}:7443/rest/v1/mkquorumapp" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"ip_6\": false, \"nometadata\": false, \"partnersystem\": \"${PARTNERSYSTEM}\", \"partnerip6\": false}"
```

Manually: Paste in the token, change the IP and partnersystem
```shell
curl -X -K 'POST' \
  'https://10.33.7.80:7443/rest/v1/mkquorumapp' \
  -H 'accept: application/json' \
  -H 'X-Auth-Token: PASTE_YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"ip_6": false, "nometadata": false, "partnersystem": "svc_cluster02", "partnerip6": false}'
```

## 3. Download the file to local folder.


The system dumps into /dumps on the virtualized config node. The API downloads it via REST and writes to your local file with --output.

```shell
curl -ks -X POST "https://${IP}:7443/rest/v1/download" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"prefix":"/dumps","filename":"ip_quorum.jar"}' \
  --output ip_quorum.jar
```

Manually: Change the IP and the Token you recived in step 1.

```shell
curl -vkX POST "https://10.33.7.80:7443/rest/v1/download" -H  "accept: application/json" -H  "X-Auth-Token: PASTE_YOUR_TOKEN" -H  "Content-Type: application/json" -d "{\"prefix\":\"/dumps\",\"filename\":\"ip_quorum.jar\"}" --output ip_quorum.jar
```

The paramter --output will "output" the file to  ip_quorum.jar

