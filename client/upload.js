import { promises as fs } from 'fs';
//import path from 'path';

async function UploadBlob(blob_path, blob) {
    try {
        const response = await fetch("http://localhost:8080/upload", {
            method: 'PUT',
            headers: { 'X-File-Path': blob_path },
            body: blob
        });
        return `Request completed with status: ${response.status}`;
    } catch (error) {
        return `Request failed: ${error}`;
    }
}

const fpath = "video.mov";
const reqpath = "videos/video.mov"
//const fpath = "thumbnail.png";
//const reqpath = "images/thumbnail.png";

const fileData = await fs.readFile(fpath);
const blob = new Blob([fileData]);
console.log(blob);

const response = await UploadBlob(reqpath, blob);
console.log(response);