const cacheName = "app-" + "eedd657c07629ef8f16e2a1ef893455d2b9e37eb";

self.addEventListener("install", event => {
  console.log("installing app worker eedd657c07629ef8f16e2a1ef893455d2b9e37eb");
  self.skipWaiting();

  event.waitUntil(
    caches.open(cacheName).then(cache => {
      return cache.addAll([
        "/pablo/test-bl",
        "/pablo/test-bl/",
        "/pablo/test-bl/app.css",
        "/pablo/test-bl/app.js",
        "/pablo/test-bl/manifest.webmanifest",
        "/pablo/test-bl/wasm_exec.js",
        "/pablo/test-bl/web/app.wasm",
        "/pablo/test-bl/web/logo192.png",
        
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
  console.log("app worker eedd657c07629ef8f16e2a1ef893455d2b9e37eb is activated");
});

self.addEventListener("fetch", event => {
  event.respondWith(
    caches.match(event.request).then(response => {
      return response || fetch(event.request);
    })
  );
});
