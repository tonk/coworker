export function lazyWithReload(importer) {
  return async () => {
    try {
      return await importer()
    } catch (err) {
      // Recover once from stale chunk URLs after deploy/update.
      if (typeof window !== 'undefined' && !sessionStorage.getItem('wd_chunk_reload_once')) {
        sessionStorage.setItem('wd_chunk_reload_once', '1')
        window.location.reload()
        return new Promise(() => {})
      }
      throw err
    }
  }
}
