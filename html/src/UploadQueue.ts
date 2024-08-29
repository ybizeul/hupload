import { APIClient } from './APIClient'
import { AxiosProgressEvent } from "axios";

export interface QueueItem {
    file : File
    loaded : number
    total : number
    finished : boolean
    failed: boolean
    error: string
}

export class UploadQueue {
    files: Record<string,QueueItem>
    path: string
    API: APIClient
    progressCallback?: (progress: QueueItem[]) => void

    constructor(api: APIClient, path: string, progress?: (progress: QueueItem[]) => void) {
        this.files = {}
        this.API = api
        this.path = path
        this.progressCallback = progress
    }
    addFiles(files: File[]) {
        files.map((f) => {
            this.files[f.name] = {
                file: f,
                loaded: 0,
                total: f.size,
                finished: false,
                failed: false,
                error:'',
            }
        })

        const promises=Object.keys(this.files).map((k) => {
            const f = this.files[k]
            const formData = new FormData();
            formData.append("file", f.file);
            
            return new Promise((resolve, reject) => {
                this.API.upload(this.path+'/items/'+f.file.name, f.file, 
                    (e: AxiosProgressEvent) => {
                        if (e.total) {
                            f.loaded = e.loaded
                            f.total = e.total
                        }
                        this.updateProgress()
                    }
                )
                .then((r) => {
                    f.finished = true; 
                    this.updateProgress()
                    resolve(r)
                })
                .catch((e) => {
                    f.finished = true; 
                    f.failed = true
                    f.error = e.response?.data.message
                    this.updateProgress()
                    reject(e)
                })
            })

        })
        return Promise.all(promises)
    }
    addFile(f: File) {
        this.addFiles([f])
    }
    updateProgress() {
        this.progressCallback?.(Object.values(this.files))
    }
}
