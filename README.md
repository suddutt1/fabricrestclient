# Generic REST API layer to access and transact with Hyperledger Fabric Networks

This generic REST API provide following functions with the  Hyperledger Fabric Networks. This should be run as a util using the command

```sh

fabricrestclient --config=<path to config json/yaml>

```
Above command exposed the REST API at port 8080 in all network interfaces of the host system.

Based on the configuration given in the json/yaml it will interact with the underlying blockchain network. Example config/yaml is available in the source. The connection profile json from IBP could also be used with minimal modification. 

## API details 
1. POST /api/chaincode/invoke
2. POST /api/chaincode/query
3. POST /api/admin/enrolladmin/:adminID
4. POST /api/admin/enrolluser
5. GET / < Service availability probe>
