
# Shock -- a data science platform

-  is a storage platform designed from the ground up to be fast, scalable and fault tolerant.

- is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

- is designed for complex scientific data and allows storage and querying of complex user-defined metadata.   

- is a data management system that supports in storage layer operations like quality-control, format conversion, filtering or subsetting.

- is integrated with S3, IBM's Tivoli TSM storage managment system.

- supports HSM operations and caching

(see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)


Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

Check out the notes  on [building and installing Shock](./building.md) and [configuration](./configuration.md).


## Shock in 30 seconds
This assumes that you have `docker` and `docker-compose` installed and `curl` is available locally.

### Download the container
`docker-compose up`

Don't forget to later `docker-compose down` and do not forget, by default this configuration does not store data persistently.

### Push a file into Shock
`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' -X PUT -F 'file_name=@myfile' http://localhost:7445/node`


### Download a file from Shock

`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' http://localhost:7445/node`

Documentation
-------------
For further information about Shock's functionality, please refer to our [Shock documentation](https://github.com/MG-RAST/Shock/docs/).
