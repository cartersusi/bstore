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


## Example
```go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	bst "github.com/cartersusi/bstore" // go pkg
    //bst "bstore" // clone
	"github.com/gin-gonic/gin"
)

func main() {
	init_file := flag.Bool("init", false, "Create a new configuration file")
	conf_file := flag.String("config", "conf.yml", "Configuration file")
	flag.Parse()
	if *init_file {
		bst.Init()
		return
	}

	bstore := &bst.ServerCfg{}
	err := bstore.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	bstore.Print()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	bstore.Cors(r)
	bstore.Middleware(r) // Must be used, able to remove/modify all middleware(rate limit, valid path) except read-write key validation

	f, err := os.OpenFile(bstore.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(f, os.Stdout)
	gin.DefaultWriter = multiWriter
	r.Use(gin.Logger())
	log.SetOutput(multiWriter)

	r.Use(bstore.Serve())
	r.PUT("/api/upload/*file_path", bstore.Upload)
	r.GET("/api/download/*file_path", bstore.Get)
	r.DELETE("/api/delete/:file_path", bstore.Delete)

	r.Run(fmt.Sprintf("%s:%s", bstore.Host, bstore.Port))
}
```

## Usage

1. Generate a Config File and Keys
```go
bstore.Init()
```

- Edit your config file (Optional)
```sh
nvim conf.yml
```

2. Clone or Import into your gin server
```sh
git clone https://github.com/cartersusi/bstore.git #clone
go get github.com/cartersusi/bstore #import
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