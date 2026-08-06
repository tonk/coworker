import client from './client'

export const webrtcApi = {
  getIceServers: () => client.get('/ice-servers'),
}
