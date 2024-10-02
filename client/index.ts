import { sleep } from 'bun';
import { promises as fs } from 'fs';

interface BlobType {
    route: string;
    method: string;
};

interface BlobStore {
    Upload: BlobType;
    Download: BlobType;
    Delete: BlobType;
};

const BStore: BlobStore = {
    Upload: { route: "upload", method: "PUT" },
    Download: { route: "download", method: "GET" },
    Delete: { route: "delete", method: "DELETE" },
};

interface TestFile {
    sourceFile: string;
    storedFile: string;
    outFile: string;
};

interface TestFiles {
    [key: string]: TestFile;
};

const test_filenames : TestFiles = {
    thumbnail: {
        sourceFile: "input/thumbnail.png",
        storedFile: "images/thumbnail.png",
        outFile: "output/thumbnail.png",
    },
    video: {
        sourceFile: "input/video.mov",
        storedFile: "videos/video.mov",
        outFile: "output/video.mov",
    },
    book: {
        sourceFile: "input/book.pdf",
        storedFile: "books/book.pdf",
        outFile: "output/book.pdf",
    },
};

async function Bstore(blob_path: string, btype: BlobType, blob?: Blob): Promise<any> {
    const full_route = `http://localhost:8080/api/${btype.route}/${blob_path}`
    console.log(`Requesting ${full_route}`);
    try {
        let response;
        if (!blob) {
            response = await fetch(full_route, {
            method: btype.method,
            });
            if (btype.method === "GET") {
                const blob = await response.blob();
                const fileData = await blob.arrayBuffer();
                return new Blob([fileData]);
            }
        } else {
            response = await fetch(full_route, {
                method: btype.method,
                body: blob,
            });
        }
        return `Request completed with status: ${response.status}`;
    } catch (error) {
        return `Request failed: ${error}`;
    }
}

async function Zstore(blob_path: string, btype: BlobType, blob?: Blob): Promise<any> {
    const full_route = `http://localhost:3000/${btype.route}/${blob_path}`
    console.log(`Requesting ${full_route}`);
    try {
        let response;
        if (!blob) {
            response = await fetch(full_route, {
            method: btype.method,
            });
            if (btype.method === "GET") {
                const blob = await response.blob();
                const fileData = await blob.arrayBuffer();
                return new Blob([fileData]);
            }
        } else {
            response = await fetch(full_route, {
                method: btype.method,
                body: blob,
            });
        }
        return `Request completed with status: ${response.status}`;
    } catch (error) {
        return `Request failed: ${error}`;
    }
}

async function uploader(x: number) {
    for (let i = 0; i < 10; i++) {
        for (const filename of Object.keys(test_filenames)) {
            const file = test_filenames[filename];
            const fileData = await fs.readFile(file.sourceFile);
            const blob = new Blob([fileData]);
            const start_time = Date.now();
            if (x === 0) {
                await Bstore(file.storedFile, BStore.Upload, blob);
            } else {
                await Zstore(file.storedFile, BStore.Upload, blob);
                }
            console.log(`Time taken for upload: ${Date.now() - start_time}ms`);
        }
    }
}

async function downloader(x: number) {
    for (let i = 0; i < 10; i++) {
        for (const filename of Object.keys(test_filenames)) {
            const file = test_filenames[filename];
            const start_time = Date.now();
            var blob;
            if (x === 0) {
                blob = await Bstore(file.storedFile, BStore.Download);
            } else {
                blob = await Zstore(file.storedFile, BStore.Download);
            }
            console.log(`Time taken for download: ${Date.now() - start_time}ms`);
            const fileData = await blob.arrayBuffer();
            await fs.writeFile(file.outFile, Buffer.from(fileData));
        }
    }
}

//uploader(0);
downloader(0);

//uploader(1);
//downloader(1);