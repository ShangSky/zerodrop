export const host = location.host;
export const httpProtocol = location.protocol;
export const wsProtocol = httpProtocol === 'https:' ? 'wss:': 'ws:'
export const uploadURL = `${httpProtocol}//${host}/upload`
export const downloadURL = `${httpProtocol}//${host}/download`
export const wsURL = `${wsProtocol}//${host}/ws`
export const fileSizeLimit = 1024 * 1024 * 10