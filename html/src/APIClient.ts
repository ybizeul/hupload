import axios, { AxiosResponse, AxiosBasicCredentials, AxiosRequestConfig, AxiosError, AxiosProgressEvent, GenericAbortSignal } from 'axios';

export const UserError = (e: unknown) => {
    if (typeof e == "string") {
        return e
    }
    if (isAPIServerError(e)) {
        return e.message
    } else {
        const er = e as AxiosError
        return (er.response?.data)?(er.response.data):(er.response?.statusText?er.response.statusText:"Unknown error")
    }
}

export interface APIServerError {
    message: string
    code: number
    response?: AxiosResponse
}

export function isAPIServerError(e: unknown): e is APIServerError { //magic happens here
    return typeof (<APIServerError>e).code == "number" && typeof (<APIServerError>e).message == "string"
}

export interface UpgradeStatusReply {
    percent: number,
    title: string,
    doneStages: string[],
    currentStage: string,
    finished: boolean,
    started: boolean,
    skipOSUpgrade: boolean,
}

export class APIClient {
    baseURL: string

    constructor(baseURL: string) {
        this.baseURL = baseURL
    }

    authFailed() {
        console.log("authFailed")
        return
    }

    logout() {
        console.log("logout")
        return
    }

    login(path: string, user? : string, password? : string) {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            axios({
                url: this.baseURL + path,
                method: 'POST',
                maxRedirects: 0,
                auth: (user && password)?{
                    username: user,
                    password: password
                }:undefined
            })
            .then((result) => {
                if (result.status == 202) {
                    resolve(result)
                } else {
                    resolve(result.data)
                }
            })
            .catch (e => {
                reject(e)
            })
        })
    }
    get(path: string, auth?: AxiosBasicCredentials) {
        return new Promise<unknown|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                method: 'GET',
                auth,
            })
            .then((result) => {
                resolve(result)
            })
            .catch(e => {
                reject(e)
            })
        })
    }

    post(path: string, data: unknown = {}) {
        return new Promise<unknown|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                method: 'POST',
                data: data
            })
            .then((result) => {
                resolve(result)
            })
            .catch (e => {
                reject(e)
            })
        })
    }

    upload(path: string, file: File, onUploadProgress?: (progressEvent: AxiosProgressEvent) => void, signal?: GenericAbortSignal): Promise<AxiosResponse|APIServerError> {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            const formData = new FormData();
            formData.append("file", file);
            
            this.request({
                url: this.baseURL + path,
                method: 'POST',
                data: formData,
                headers: {
                    'Content-Type': 'multipart/form-data',
                    'FileSize': file.size.toString(),
                },
                onUploadProgress,
                signal,
            })
            .then((result) => {
                resolve(result)
            })
            .catch(reject)
        })
    }

    patch(path: string, data = {}) {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                method: 'PATCH',
                data: data
            })
            .then((result) => {
                resolve(result)
            })
            .catch(reject)
        })
    }

    delete(path: string, data={}) {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                method: 'DELETE',
                data: data,
            })
            .then((result) => {
                resolve(result)
            })
            .catch (reject)
        })
    }

    put(path: string, data = {}) {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                method: 'PUT',
                data: data,
            })
            .then((result) => {
                resolve(result)
            })
            .catch (reject)
        })
    }

    wait(s=0) {
        return new Promise((resolve, reject) => {
            // eslint-disable-next-line @typescript-eslint/no-this-alias
            const n = this
            let cancel = false
            //const abort = new AbortController()
            if (s>0) {
                new Promise(r => setTimeout(r, s*1000))
                .then(() => {
                    cancel = true
                })
            }

            this.get('/health')
            .then(() => {
                resolve(true)
            })
            .catch(function retry(e) {
                if (e.code == "ERR_NETWORK" || e.code == "ERR_BAD_RESPONSE" || e.code == "ECONNABORTED") {
                    return new Promise(r => setTimeout(r, 2000))
                    .then(() => {
                        return n.get('/health')
                        .then(() => resolve(true))
                        .catch((e) => {
                            if (cancel) {
                                reject("Timeout")
                            } else {
                                retry(e)
                            }
                        })
                    })
                }
                else {
                    reject(e)
                }
            })
        })
    }

    waitDown(s=0) {
        return new Promise((resolve, reject) => {
            // eslint-disable-next-line @typescript-eslint/no-this-alias
            const n = this
            let cancel = false
            //const abort = new AbortController()
            if (s>0) {
                new Promise(r => setTimeout(r, s*1000))
                .then(() => {
                    cancel = true
                })
            }
            this.get('/health')
            .then(function retry() {
                return new Promise(r => setTimeout(r, 2000))
                    .then(() => {
                        return n.get('/health')
                        .then(() => {
                            if (cancel) {
                                reject("Timeout")
                            } else {
                                retry()
                            }
                        })
                        .catch(() => {
                            resolve(true)
                        })
                    })
            })
            .catch(() => {
                resolve(true)
            })
        })
    }

    config(path: string, c: AxiosRequestConfig) {
        return new Promise<AxiosResponse|APIServerError>((resolve, reject) => {
            this.request({
                url: this.baseURL + path,
                ...c
            })
            .then((result) => {
                resolve(result)
            })
            .catch((e) => {
                reject(e)
            })
        })
    }

    request(config: AxiosRequestConfig) {
        return new Promise<AxiosResponse>((resolve, reject) => {
            // Send request with regular token
            //config.headers = {"Authorization": "Bearer " + window.localStorage.getItem('token')}
            axios(config)
            .then((result) => {
                resolve(result.data)
            })
            .catch((e) => {
                const ae = e as AxiosError
                if (ae.response) {
                    // We have a live server,
                    // Check if we need to refresh token
                    if (ae.response.status == 401 && ae.response?.headers["www-authenticate"]) {
                        this.authFailed()
                        // Regular token failed, config for refresh token
                        //window.location.reload()
                        reject(ae)
                        return
                        //config.headers = {"Authorization": "Bearer " + window.localStorage.getItem('refresh')}
                    }
                    else {
                        // It's not an authentication problem, handle the exception
                        if (ae.response.data && isAPIServerError(ae.response.data)) {
                            // There is data, and that's an exception
                            reject(ae)
                        } else {
                            // It's a generic exception, return as-is
                            reject(ae)
                            return
                        }
                    }
                } else {
                    // There is no data in the error, return as-is
                    console.log(e)
                    reject(e)
                    return
                }
            })
        })
    }
}

class HuploadClient extends APIClient {
    constructor() {
        super('/api/v1')
    }
    // authFailed() {
    //     if (!window.location.href.endsWith("/login")) {
    //         window.location.href = "/login"
    //     }
    // }
    logoutNow() {
        document.cookie = "X-Token" +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        document.cookie = "X-Token-Refresh" +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    }
}

export const H = new HuploadClient()
