# Bstore
This code is not meant for production. Only posted as reference. If you like the idea and want to help create a complete package feel free to reach out.

## About 
A simple and fast file server for serving blob files.
* Secure: AES 256-bit encryption
* Efficient: zstd compression

## Use Cases
* DIY Movies/TV Server
* PDF Books
* Data Backups


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

# API Usage

## Upload a File
- Fetch `/api/upload/*` to upload a file
```js
const readWriteKey = process.env.BSTORE_READ_WRITE_KEY || '';
var file_path = 'cats/siamese/cat.png'

const file = Bun.file(file_path);

const res = await fetch(`https://catlovers.com/api/upload/${file_path}`, {
    method: 'PUT',
    headers: {
        'X-access': 'public',
        'Authorization': `Bearer ${readWriteKey}`,
        'Content-Type': file.type,
      },
    body: file,
});
```

## Download a File
- Fetch `/api/download/*` to download a file
```js
const readWriteKey = process.env.BSTORE_READ_WRITE_KEY || '';

const res = await fetch(`https://catlovers.com/api/download/${file_path}`, {
    method: 'GET',
    headers: {
        'X-access': 'public',
        'Authorization': `Bearer ${readWriteKey}`,
    },
});
const blob = await res.blob();
await Bun.write("cat_copy.png", blob);
```

## Delete a File
- Fetch `/api/delete/*` to delete a file
```js
const readWriteKey = process.env.BSTORE_READ_WRITE_KEY || '';

const res = await fetch(`https://catlovers.com/api/delete/${file_path}`, {
    method: 'DELETE',
    headers: {
        'X-access': 'public',
        'Authorization': `Bearer ${readWriteKey}`,
    },
});
```

---

# Serve a File
```html
<img src="https://catlovers.com/bstore/cats/siamese/cat.png" alt="Siamese Cat" className="max-w-full max-h-full object-contain" />
<video src="https://catlovers.com/bstore/homepage/cats.mp4" controls className="max-w-full max-h-full" />
<embed src="https://catlovers.com/bstore/info/cats.pdf" type="application/pdf" width="100%" height="600px" />
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