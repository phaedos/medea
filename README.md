# Medea - A sufficient and simple data transport and sync system.

In the real-time digital world, accomplishing high performance along the entire transfer path requires a new and fundamentally different approach to bulk data movement.

Medea is a data transfer and sync system. By providing sufficient file management services and an efficient transfer framework, medea provides a high-performance and scalable solution for cross-regional data synchronization tasks.

# Features
Supported features:
- HTTP/HTTPS protocol support
  + Support cors
  + Support to avoid replay attack
  + Support to validate parameter signature
- range download

Upcoming features:
- out-of-order upload
- Fragment md5 checksum
- QUIC transport
- database model extension

# Quick start

The usecase of medea is divided into four steps：
- step0: build
- step1: prepare mysql database environment
- step2: create app
- step3: upload and download file

## step0 build
```
make
```

## step1 prepare mysql

Make sure the mysql database is accessible

## step2: create app

1 start server

```
./medea --config ./deploy/medea.yaml http:start
```

2 create app

```
./medea app:new --name example
```

3 list app

```
./medea app:list
```

## step3 upload and download file

1 fetch client token

```
./medea client:token --uid c180f6c861b4eb900b4948855b3ea40d --secret BN20TZKDzB6W
```


2 upload file

```
./medea client:file --token 986403d6e2358ffe5add741c693f485f --secret 1a17a12d604f404ecc9363cdc3457521 --path /example/test --src ./test
```

3 download file

```
./medea client:read --token 986403d6e2358ffe5add741c693f485f --secret 1a17a12d604f404ecc9363cdc3457521 --uid bf9edce9f68441c6a878ee35577cf673 --dst ./test_local
```

4 download in range mode
```
./medea client:read --token 986403d6e2358ffe5add741c693f485f --secret 1a17a12d604f404ecc9363cdc3457521 --uid bf9edce9f68441c6a878ee35577cf673 --dst ./test_local --range 4
```

# Notice

Environment general information could be configured before client operations.
```
./medea client:env --server http://127.0.0.1:8630 --host 172.0.0.1
```
