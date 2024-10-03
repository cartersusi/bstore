# Bstore
A simple and fast file server for serving blob files.
* Fast: 
* Secure: 
* Efficient: Uses zstd compression for

## Usage
1. Clone the repo
```sh
git clone https://github.com/cartersusi/bstore.git
```

2. Build bstore
```sh
cd bstore
make build
```

3. Create a read/write key
```sh
export BSTORE_READ_WRITE_KEY=$(openssl rand -base64 32)
echo $BSTORE_READ_WRITE_KEY # copy for later
```

4. Generate a config File
```sh
./bstore -init
```

5. (Optional) Edit your config file
```sh
nvim conf.yml
```

6. Run the server
```sh
./bstore
```

---

# API Usage

```js
const readWriteKey = process.env.BSTORE_READ_WRITE_KEY || '';
var file_path = 'cats/siamese/cat.png'
```

## Upload a File
- Fetch `/api/upload/*` to upload a file
```js
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
const res = await fetch(`https://catlovers.com/api/download/${file_path}`, {
    method: 'GET',
    headers: {
        'X-access': 'public',
        'Authorization': `Bearer ${readWriteKey}`,
    },
});
const blob = await res.blob();
await Bun.write("cat_copy.webp", blob);
```

## Delete a File
- Fetch `/api/delete/*` to delete a file
```js
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