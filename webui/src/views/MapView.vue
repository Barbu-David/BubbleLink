<!-- src/views/MapView.vue -->
<template>
  <div class="page">
    <div id="map" ref="mapEl"></div>

    <aside class="panel" :class="{ open: sidebarOpen }">
      <header>
        <h2>Add a Place</h2>
        <button class="ghost" @click="sidebarOpen = false">Close</button>
      </header>

      <div v-if="draft">
        <div class="row">
          <label>Title</label>
          <input v-model.trim="draft.title" placeholder="e.g. Hidden courtyard" />
        </div>
        <div class="row">
          <label>Notes</label>
          <textarea v-model.trim="draft.notes" rows="4" placeholder="Why is this spot special?"></textarea>
        </div>
        <div class="meta">
          <span>Lat: {{ draft.lat.toFixed(5) }}</span>
          <span>Lng: {{ draft.lng.toFixed(5) }}</span>
        </div>
        <button class="cta" @click="saveDraft">Save (local)</button>
      </div>

      <div v-else class="muted">Drop a marker from the map toolbar to start.</div>

      <hr />

      <h3>Your Local Places</h3>
      <div v-if="places.length === 0" class="muted">No places yet.</div>
      <ul class="list">
        <li v-for="p in places" :key="p.id" @click="goTo(p)">
          <div class="dot"></div>
          <div class="txt">
            <div class="t">{{ p.title || 'Untitled place' }}</div>
            <div class="s">({{ p.lat.toFixed(5) }}, {{ p.lng.toFixed(5) }})</div>
          </div>
        </li>
      </ul>
    </aside>

    <div class="toolbar">
      <button title="Add place at center" @click="addAtCenter">＋</button>
      <button title="Geolocate" @click="locate">📍</button>
      <button title="Reset view" @click="resetView">⤾</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// Fix default marker icons in Vite
import iconUrl from 'leaflet/dist/images/marker-icon.png'
import iconRetinaUrl from 'leaflet/dist/images/marker-icon-2x.png'
import shadowUrl from 'leaflet/dist/images/marker-shadow.png'
L.Icon.Default.mergeOptions({ iconUrl, iconRetinaUrl, shadowUrl })

const mapEl = ref(null)
let map
let centerMarker = null

const sidebarOpen = ref(false)
const draft = ref(null)
const places = ref([])

const DEFAULT = { lat: 44.439663, lng: 26.096306, zoom: 13 } // Bucharest default

function resetView() {
  map.setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
}

function addAtCenter() {
  const c = map.getCenter()
  const marker = L.marker(c, { draggable: true }).addTo(map)
  if (centerMarker) map.removeLayer(centerMarker)
  centerMarker = marker

  draft.value = { id: crypto.randomUUID(), lat: c.lat, lng: c.lng, title: '', notes: '' }
  sidebarOpen.value = true

  marker.on('drag', (ev) => {
    draft.value.lat = ev.latlng.lat
    draft.value.lng = ev.latlng.lng
  })
}

function saveDraft() {
  if (!draft.value) return
  places.value.unshift({ ...draft.value })
  draft.value = null
  sidebarOpen.value = false
  if (centerMarker) {
    centerMarker.bindPopup('Saved locally').openPopup()
    centerMarker = null
  }
}

function goTo(p) {
  map.flyTo([p.lat, p.lng], 17, { duration: 0.8 })
}

function locate() {
  map.locate({ setView: true, maxZoom: 16 })
}

onMounted(() => {
  map = L.map(mapEl.value, { zoomControl: false }).setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map)

  // controls
  L.control.zoom({ position: 'bottomright' }).addTo(map)

  // click to preview coordinates
  map.on('click', (e) => {
    L.popup().setLatLng(e.latlng).setContent(`Lat ${e.latlng.lat.toFixed(5)}, Lng ${e.latlng.lng.toFixed(5)}`).openOn(map)
  })

  window.addEventListener('keydown', onKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  map && map.remove()
})

function onKey(e) {
  if (e.key.toLowerCase() === 'a') addAtCenter()
  if (e.key.toLowerCase() === 'l') locate()
}
</script>

<style scoped>
.page { position: relative; height: calc(100vh - 56px); }
#map { position: absolute; inset: 0; border-top: 1px solid rgba(255,255,255,0.05); }
.toolbar {
  position: absolute; left: 16px; bottom: 16px; z-index: 500;
  display: flex; flex-direction: column; gap: 8px;
}
.toolbar button {
  width: 44px; height: 44px; border-radius: 12px; border: 1px solid rgba(255,255,255,.15);
  background: rgba(0,0,0,.35); color: #fff; font-size: 20px; cursor: pointer;
  box-shadow: 0 6px 24px rgba(0,0,0,.25);
}
.panel {
  position: absolute; right: 0; top: 0; bottom: 0; width: 360px; transform: translateX(100%);
  transition: transform .25s ease; z-index: 600;
  background: rgba(10,12,20,.9); border-left: 1px solid rgba(255,255,255,.08); padding: 14px;
  color: #e9ecf1;
}
.panel.open { transform: translateX(0); }
.panel header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.ghost { background: transparent; border: 1px solid rgba(255,255,255,.2); color: #cfe6ff; padding: 6px 10px; border-radius: 8px; cursor: pointer; }
.row { display: grid; gap: 8px; margin: 10px 0; }
input, textarea {
  width: 100%; padding: 10px 12px; border-radius: 10px; border: 1px solid rgba(255,255,255,.15);
  background: rgba(255,255,255,.06); color: #fff;
}
.meta { display: flex; gap: 12px; font-size: 12px; opacity: .7; margin-bottom: 10px; }
.cta {
  width: 100%; display: grid; place-items: center;
  background: linear-gradient(135deg, #ffd36a, #ff8e6e);
  color: #0c0f1a; font-weight: 800; border: 0; border-radius: 12px; padding: 12px; cursor: pointer;
}
hr { border: 0; height: 1px; background: rgba(255,255,255,.08); margin: 14px 0; }
.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 8px; }
.list li {
  display: grid; grid-template-columns: 12px 1fr; gap: 10px; align-items: center;
  padding: 8px; border-radius: 10px; cursor: pointer; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06);
}
.dot { width: 8px; height: 8px; border-radius: 50%; background: #8bd0ff; }
.txt .t { font-weight: 700; }
.txt .s { font-size: 12px; opacity: .65; }
.muted { opacity: .7; }
</style>
