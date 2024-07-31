import { APIClient } from './APIClient'
import { AxiosProgressEvent } from "axios";

export interface QueueItem {
    file : File
    loaded : number
    total : number
    finished : boolean
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
            }
        })
        Object.keys(this.files).map((k) => {
            const f = this.files[k]
            const formData = new FormData();
            formData.append("file", f.file);
            // axios.request({
            //     url: '/api/v1/' + this.path + '/' + f.file.name,
            //     method: 'POST',
            //     data: formData,
            //     headers: {
            //         'Content-Type': 'multipart/form-data'
            //     },
            //     onUploadProgress: (e: AxiosProgressEvent) => {
            //         console.log(e)
            //         if (e.total) {
            //             f.loaded = e.loaded
            //             f.total = e.total
            //         }
            //         this.updateProgress()
            //     }
            // })
            this.API.upload(this.path+'/'+f.file.name, f.file, 
                (e: AxiosProgressEvent) => {
                    console.log(e)
                    if (e.total) {
                        f.loaded = e.loaded
                        f.total = e.total
                    }
                    this.updateProgress()
                }
            )
            .then(() => {f.finished = true; this.updateProgress()})
            .catch((e) => {console.log(e)})
        })
    }
    addFile(f: File) {
        this.addFiles([f])
    }
    updateProgress() {
        this.progressCallback?.(Object.values(this.files))
    }
}
