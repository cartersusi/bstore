import { promises as fs } from 'fs';
//import path from 'path';

async function GetBlob(blob_path) {
    try {
        const response = await fetch("http://localhost:8080/get", {
            method: 'GET',
            headers: { 'X-File-Path': blob_path }
        });
        console.log(`Request completed with status: ${response.error}`);

        const blob = await response.blob();
        const fileData = await blob.arrayBuffer();
        return new Blob([fileData]);
    } catch (error) {
        return `Request failed: ${error}`;
    }
}

const fpath = "videos/video.mov";
const outputPath = "returned_video.mov";
//const fpath = "images/thumbnail.png";
//const outputPath = "returned_thumbnail.png";

const blob = await GetBlob(fpath);
const fileData = await blob.arrayBuffer();
await fs.writeFile(outputPath, Buffer.from(fileData));