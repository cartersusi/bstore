<div align="center">
  <img width="192px" height="auto" src="public/favicon.ico" alt="Bstore Logo">
  <h1>Bstore</h1>
  <p>A simple blob storage.</p>
</div>

<div align="center">

  [![Go Package](https://img.shields.io/badge/go%20package-bstore-00ADD8?style=flat-square&logo=go)](https://github.com/carterusi/bstore)
  [![NPM Package](https://img.shields.io/badge/npm-bstorejs-red?style=flat-square&logo=npm)](https://www.npmjs.com/package/bstorejs)
  [![React Package](https://img.shields.io/badge/react-bstorejs--react-61DAFB?style=flat-square&logo=react)](https://www.npmjs.com/package/bstorejs-react)
  [![Demo](https://img.shields.io/badge/demo-bstorejs--demo-brightgreen?style=flat-square)](https://github.com/carterusi/bstorejs-demo)

</div>

## About 
* Secure: AES 256-bit encryption
* Efficient: zstd compression

## Use Cases
* DIY Movies/TV Server
* PDF Books
* Data Backups


## APIs
- [bstorejs](https://github.com/cartersusi/bstorejs.git) - Express/Vanilla Js/Ts APIs
```sh
npm i bstorejs
```
- [bstorejs-react](https://github.com/cartersusi/bstorejs-react.git) - React Server Actions & Components
```sh
npm i bstorejs-react
```

## Usage

1. **Clone Repository**
```sh
git clone https://github.com/cartersusi/bstore.git
```

2. **Build For your OS**
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
nvim conf.yml
```

4. **Start Server**
```sh
./bstore
```

- **Use a different config**
```sh
./bstore -config new_conf.yml
```

---
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