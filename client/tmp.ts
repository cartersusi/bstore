import { promises as fs } from 'fs';

const fileData = await fs.readFile('new.pdf');
const blob = new Blob([fileData]);

const uploadresponse = await fetch("http://0.0.0.0:3000/upload/books/large_book.pdf", {
    method: "PUT",
    body: blob,
});

console.log(uploadresponse.status);