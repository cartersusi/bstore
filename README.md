<div align="center">
  <img width="192px" height="auto" src="support/favicon.ico" alt="Bstore Logo">
  <h1>Bstore</h1>
  <p>A simple blob storage.</p>
</div>

<div align="center">

  [![bstore](https://img.shields.io/badge/go-bstore-00ADD8?style=flat-square&logo=go)](https://github.com/cartersusi/bstore)
  [![NPM Package](https://img.shields.io/badge/npm-bstorejs-red?style=flat-square&logo=npm)](https://www.npmjs.com/package/bstorejs)
  [![React Package](https://img.shields.io/badge/react-bstorejs--react-61DAFB?style=flat-square&logo=react)](https://www.npmjs.com/package/bstorejs-react)
  [![Demo](https://img.shields.io/badge/demo-bstorejs--demo-brightgreen?style=flat-square)](https://github.com/cartersusi/bstore-demo)

</div>

## About 
### **Fast**: 
**1mb** files **encrypted** and **compressed**.

|Storage|Tier|Upload|Download|
|-|-|-|-|
[bstorejs](https://www.npmjs.com/package/bstorejs) | 8c/16t - 1 Gbps | 924 upload/s | 617 download/s|
[@vecel/blob](https://www.npmjs.com/package/@vercel/blob)| Free | 4.1 upload/s | 57 download/s |
[@aws-sdk/client-s3](https://www.npmjs.com/package/@aws-sdk/client-s3)| Free | 5.3 upload/s | 72 download/s |

### **Secure**: 
  * AES 256-bit encryption

### **Efficient**: 
  * zstd compression

### Lightweight:
|OS|Size|
|-|-|
|darwin-amd64|(8.44 MB)|
|darwin-arm64|(8.09 MB)|
|linux-arm64|(7.94 MB)|
|linux-amd64|(8.25 MB)|

## Use Cases
* DIY Movies/TV Server
* PDF Books
* Data Backups

## Features 
* HLS and MPEG DASH video Streaming
* Data Cache
* Rate Limiting

## Build (Recommended)

1. **Clone Repository**
```sh
git clone https://github.com/cartersusi/bstore.git
```

2. **Build For your OS (Requires Go)**
```sh
cd bstore
make build
```

3. **Generate a Config File and Keys**
```go
./bstore -init
```

- **Edit your config file (Optional)**
  ```sh
  nvim ~/.bstore/conf.yml
  ```

- **Print your keys (Optional)**
  ```sh
  cat ~/.bstore/keys.env
  ```

4. **Start Server**
```sh
./bstore
```

- **Use a different config (Optional)**
  ```sh
  ./bstore -config new_conf.yml
  ```

## Install
```sh
curl -fsSL https://cartersusi.com/bstore/install | bash
```

## APIs
- [bstorejs](https://github.com/cartersusi/bstorejs.git) - Express/Vanilla Js/Ts APIs
```sh
npm i bstorejs
```
- [bstorejs-react](https://github.com/cartersusi/bstorejs-react.git) - React Server Actions & Components
```sh
npm i bstorejs-react
```
<!-- 
# `bstore` npm package

## Upload a File
```ts
import { put, PutBstoreResponse } from 'bstore';

// upload a public file
const res: PutBstoreResponse = await put(file, file_path, 'public');
//upload a private file
const res: PutBstoreResponse = await put(file, file_path, 'private');
```

## Download a File
```ts
import {get, GetBstoreResponse} from 'bstore';

// download a public file
const res: GetBstoreResponse = await get("/images/image.png", 'public');
//download a private file
const res: GetBstoreResponse = await get("/books/book.pdf", 'private');
```

## Delete a File
```ts
import {del, DeleteBstoreResponse} from 'bstore';

// delete a file
const res: DeleteBstoreResponse = await del("/books/book.pdf", 'public');
// delete a directory
const res: DeleteBstoreResponse = await del("/images/hentai/*", 'private');
```

-->
