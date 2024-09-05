// Interfaces

export interface ShareOptions {
  exposure?: string;
  validity?: number;
  description?: string;
  message?: string;
}

export interface Share {
  name: string;
  owner: string;
  count: number;
  size: number;
  created: Date;
  options: ShareOptions;
  isvalid: boolean;
}

export interface Item {
  Path: string;
  ItemInfo: ItemInfo;
}

export interface ItemInfo {
    Size: number;
}

export interface Message {
  title: string;
  message: string;
}

// Utilities

export function humanFileSize(bytes: number, si=false, dp=1) {
    const thresh = si ? 1000 : 1024;
  
    if (Math.abs(bytes) < thresh) {
      return bytes + ' B';
    }
  
    const units = si 
      ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'] 
      : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10**dp;
  
    do {
      bytes /= thresh;
      ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);
  
  
    return bytes.toFixed(dp) + ' ' + units[u];
  }

  export function prettyfiedCount(count: number|null, singular: string, plural: string, empty: string|null) {
    if (count === 0 || count === null|| count === undefined) {
      return empty
    } else if (count === 1) {
      return count.toFixed() + ' ' + singular
    } else {
      return count.toFixed() + ' ' + plural
    }
  }