const readWriteKey = process.env.BSTORE_READ_WRITE_KEY || '';
const bstoreHost = process.env.BSTORE_HOST || 'http://localhost:8080';

export enum Method {
    GET = 0,
    PUT,
    DELETE,
    LIST,
}

const Methods = {
    [Method.GET]: {
        route: 'api/download/',
        method: 'GET'
    },
    [Method.PUT]: {
        route: 'api/upload/',
        method: 'PUT'
    },
    [Method.DELETE]: {
        route: 'api/delete/', 
        method: 'DELETE'
    },
    [Method.LIST]: {
        route: 'api/list/',
        method: 'GET'
    }
}

export interface BstoreRequest {
    method: Method;
    path: string;
    access: 'public' | 'private';
    file?: File;
}

export interface BstoreResponse {
    status: number;
    message: string;
    url?: string;
    file?: File;
    files?: string[];
}

interface BstoreHeaders {
    [key: string]: string;
}

export async function bstore({ method, path, access, file }:BstoreRequest): Promise<BstoreResponse> {
    var ret: BstoreResponse = {
        status: 0,
        message: 'Failed to fetch'
    }

    var headers: BstoreHeaders = {
        'X-access': access,
        'Authorization': `Bearer ${readWriteKey}`,
    }
    if (file) {
        headers['Content-Type'] = file.type;
    }

    var api_url = `${bstoreHost}/${Methods[method].route}/${path}`;
    api_url = api_url.replace(/\/\//g, '/');

    const res = await fetch(api_url, {
        method: Methods[method].method,
        headers: headers,
        body: file,
    });
    
    if (res.status !== 200) {
        ret.status = res.status;
        ret.message = res.statusText;
        return ret;
    }

    ret.status = res.status;
    ret.message = `Successful: ${Methods[method].method} ${path}`;

    if (method === Method.GET) {
        const blob = await res.blob();
        const file = new File([blob], path.split('/').pop() || 'file', { type: blob.type });
        ret.file = file;
    }

    if (method === Method.PUT) {
        const url = await res.text();
        ret.url = url;
    }

    if (method === Method.LIST) {
        const files = await res.json();
        ret.files = files;
    }

    return ret;
}

const res: BstoreResponse = await bstore({
    method: Method.LIST,
    path: '',
    access: 'public',
});

console.log(res);