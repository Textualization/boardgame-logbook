const cacheName = "app-" + "c7a21deabfa747de3c94c3d8087e62ba57c0ad9c";

self.addEventListener("install", event => {
  console.log("installing app worker c7a21deabfa747de3c94c3d8087e62ba57c0ad9c");
  self.skipWaiting();

  event.waitUntil(
    caches.open(cacheName).then(cache => {
      return cache.addAll([
        "/boardgame-logbook",
        "/boardgame-logbook/",
        "/boardgame-logbook/app.css",
        "/boardgame-logbook/app.js",
        "/boardgame-logbook/manifest.webmanifest",
        "/boardgame-logbook/wasm_exec.js",
        "/boardgame-logbook/web/app.wasm",
        "/boardgame-logbook/web/logo192.png",
        
      ]);
    })
  );
});

self.addEventListener("activate", event => {
  event.waitUntil(
    caches.keys().then(keyList => {
      return Promise.all(
        keyList.map(key => {
          if (key !== cacheName) {
            return caches.delete(key);
          }
        })
      );
    })
  );
  console.log("app worker c7a21deabfa747de3c94c3d8087e62ba57c0ad9c is activated");
});

self.addEventListener("fetch", event => {
  event.respondWith(
    caches.match(event.request).then(response => {
      return response || fetch(event.request);
    })
  );
});
