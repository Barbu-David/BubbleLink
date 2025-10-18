<!-- src/views/MapView.vue -->
<template>
  <div class="layout">
    <!-- Map -->
    <div id="map" ref="mapEl"></div>

    <!-- Permanent sidebar -->
    <aside class="panel">
      <header>
        <h2>Places</h2>
        <div class="muted small">
          <span v-if="zoom < ZOOM_MIN">Zoom in to load places (min {{ ZOOM_MIN }})</span>
          <span v-else>Showing {{ visiblePlaces.length }} / {{ allPlaces.length }}</span>
        </div>
      </header>

      <!-- Draft creator -->
      <section class="card">
        <h3>Add a Place</h3>
        <div v-if="draft">
          <div class="row">
            <label>Title</label>
            <input v-model.trim="draft.title" placeholder="e.g. Hidden courtyard" />
          </div>
          <div class="row">
            <label>Notes</label>
            <textarea v-model.trim="draft.notes" rows="3" placeholder="Why is this spot special?"></textarea>
          </div>
          <div class="meta">
            <span>Lat: {{ draft.lat.toFixed(5) }}</span>
            <span>Lng: {{ draft.lng.toFixed(5) }}</span>
          </div>
          <button class="cta" @click="saveDraft">Save (local)</button>
        </div>
        <div v-else class="muted">Use the ＋ button to add at map center.</div>
      </section>

      <hr />

      <!-- Visible places list -->
      <h3>Your Places (visible)</h3>
      <div v-if="zoom >= ZOOM_MIN">
        <div v-if="visiblePlaces.length === 0" class="muted">No places in view.</div>
        <ul class="list">
          <li v-for="p in visiblePlaces" :key="p.id" @click="openPlace(p)">
            <div class="dot"></div>
            <div class="txt">
              <div class="t">{{ p.title || 'Untitled place' }}</div>
              <div class="s">({{ p.lat.toFixed(5) }}, {{ p.lng.toFixed(5) }})</div>
            </div>
          </li>
        </ul>
      </div>
      <div v-else class="muted">Zoom in to at least {{ ZOOM_MIN }} to load markers.</div>

      <hr />

      <!-- Tools -->
      <div class="toolbar">
        <button title="Add place at center" @click="addAtCenter">＋</button>
        <button title="Geolocate" @click="locate">📍</button>
        <button title="Reset view" @click="resetView">⤾</button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, computed } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { useAuth } from '@/composables/useAuth' // ⬅️ use same composable as App.vue

// Fix Leaflet default png icons (only for the temporary draggable draft marker)
import iconUrl from 'leaflet/dist/images/marker-icon.png'
import iconRetinaUrl from 'leaflet/dist/images/marker-icon-2x.png'
import shadowUrl from 'leaflet/dist/images/marker-shadow.png'
L.Icon.Default.mergeOptions({ iconUrl, iconRetinaUrl, shadowUrl })

// auth
const { user } = useAuth() // user is a ref, e.g. { userId, name, city, country, ... }

const mapEl = ref(null)
let map
let centerMarker = null
let markersLayer        // LayerGroup for saved places
let homeLayer           // LayerGroup for "home" pin (kept separate so we don't clear it)
const userHomeCoords = ref(null) // [lat, lon] once geocoded, for instant reset

const draft = ref(null)
const allPlaces = ref([])          // all saved places (local)
const zoom = ref(13)
const bounds = ref(null)

const DEFAULT = { lat: 44.439663, lng: 26.096306, zoom: 13 } // fallback view only
const ZOOM_MIN = 13

// ------ Helpers to read user's base city/country safely ------
const userBase = computed(() => {
  const u = user?.value
  const city = (u && typeof u.city === 'string' && u.city.trim()) ? u.city.trim() : null
  const country = (u && typeof u.country === 'string' && u.country.trim()) ? u.country.trim() : null
  return (city && country) ? { city, country } : null
})
// -------------------------------------------------------------

function resetView() {
  // Instant reset to user's geocoded home if we have it:
  if (userHomeCoords.value) {
    map.setView(userHomeCoords.value, Math.max(DEFAULT.zoom, 12))
    return
  }
  // Otherwise, try to center to user's base (async geocode). Fallback to DEFAULT.
  if (userBase.value) {
    centerOnUserBase() // fire-and-forget
  } else {
    map.setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
  }
}

function addAtCenter() {
  const c = map.getCenter()
  if (centerMarker) map.removeLayer(centerMarker)
  centerMarker = L.marker(c, { draggable: true }).addTo(map)

  draft.value = { id: crypto.randomUUID(), lat: c.lat, lng: c.lng, title: '', notes: '' }

  centerMarker.on('drag', (ev) => {
    draft.value.lat = ev.latlng.lat
    draft.value.lng = ev.latlng.lng
  })
}

function saveDraft() {
  if (!draft.value) return
  const place = { ...draft.value }
  allPlaces.value.unshift(place)
  persistLocal()
  addMarkerForPlaceIfVisible(place)
  savePlaceToDB(place).catch(() => {})

  draft.value = null
  if (centerMarker) {
    map.removeLayer(centerMarker)
    centerMarker = null
  }
}

function openPlace(p) {
  map.flyTo([p.lat, p.lng], Math.max(zoom.value, 17), { duration: 0.8 })
  markersLayer.eachLayer(layer => {
    if (layer._placeId === p.id) layer.openPopup()
  })
}

function locate() {
  map.locate({ setView: true, maxZoom: 16 })
}

function savePlaceToDB(place) {
  // TODO: replace with real API call
  return Promise.resolve()
}

// Visible places (zoom + bounds filtered)
const visiblePlaces = computed(() => {
  if (!bounds.value || zoom.value < ZOOM_MIN) return []
  const b = bounds.value
  return allPlaces.value.filter(p => b.contains([p.lat, p.lng]))
})

// ---------- Geocoding: city+country -> center map & add "home" pin ----------
function getCacheKey(city, country) {
  return `geocode:${(city||'').toLowerCase()}|${(country||'').toLowerCase()}`
}

async function geocodeCityCountry(city, country) {
  const key = getCacheKey(city, country)
  const cached = localStorage.getItem(key)
  if (cached) {
    try { return JSON.parse(cached) } catch {}
  }

  const q = encodeURIComponent(`${city}, ${country}`)
  const url = `https://nominatim.openstreetmap.org/search?q=${q}&format=json&limit=1&addressdetails=0`

  const res = await fetch(url, { headers: { 'Accept-Language': 'en' } })
  if (!res.ok) throw new Error('Geocoding failed')
  const arr = await res.json()
  if (!arr.length) throw new Error('No results for that city/country')

  const first = arr[0]
  // Nominatim /search returns boundingbox as [south, north, west, east]
  const south = parseFloat(first.boundingbox?.[0])
  const north = parseFloat(first.boundingbox?.[1])
  const west  = parseFloat(first.boundingbox?.[2])
  const east  = parseFloat(first.boundingbox?.[3])

  const result = {
    lat: parseFloat(first.lat),
    lon: parseFloat(first.lon),
    bbox: (Number.isFinite(south) && Number.isFinite(north) && Number.isFinite(west) && Number.isFinite(east))
      ? [south, west, north, east] : null,
    displayName: first.display_name
  }
  localStorage.setItem(key, JSON.stringify(result))
  return result
}

const HOME_ICON = L.divIcon({
  className: 'pin-home',
  iconSize: [20, 20],
  iconAnchor: [10, 20],
  popupAnchor: [0, -18],
  html: '<div class="pin-head"></div><div class="pin-stick"></div>'
})

async function centerOnUserBase() {
  try {
    const base = userBase.value
    if (!base) {
      map.setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
      return
    }
    const geo = await geocodeCityCountry(base.city, base.country)
    userHomeCoords.value = [geo.lat, geo.lon]

    // draw (or replace) home pin
    homeLayer.clearLayers()
    L.marker(userHomeCoords.value, { icon: HOME_ICON })
      .bindTooltip(`${base.city}, ${base.country}`)
      .bindPopup(`<strong>${escapeHTML(base.city)}, ${escapeHTML(base.country)}</strong>`)
      .addTo(homeLayer)

    // Fit to bbox if available, else center
    if (geo.bbox) {
      const [south, west, north, east] = geo.bbox
      map.fitBounds(L.latLngBounds([south, west], [north, east]), { padding: [20, 20] })
    } else {
      map.setView(userHomeCoords.value, Math.max(DEFAULT.zoom, 12))
    }
  } catch (e) {
    console.warn('centerOnUserBase failed; using fallback', e)
    map.setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
  }
}
// ---------------------------------------------------------------------------

// Render helpers for saved places
function createRedDivIcon() {
  return L.divIcon({
    className: 'pin-red',
    iconSize: [18, 18],
    iconAnchor: [9, 18],
    popupAnchor: [0, -18],
    html: '<div class="pin-head"></div><div class="pin-stick"></div>'
  })
}

function addMarkerForPlace(place) {
  const m = L.marker([place.lat, place.lng], { icon: createRedDivIcon() })
    .bindTooltip(place.title || 'Untitled place')
    .bindPopup(`<strong>${escapeHTML(place.title || 'Untitled place')}</strong><br>${escapeHTML(place.notes || '')}`)
  m._placeId = place.id
  m.addTo(markersLayer)
}

function addMarkerForPlaceIfVisible(place) {
  if (zoom.value >= ZOOM_MIN && bounds.value && bounds.value.contains([place.lat, place.lng])) {
    addMarkerForPlace(place)
  }
}

function renderVisibleMarkers() {
  if (!markersLayer) return
  markersLayer.clearLayers()
  if (zoom.value < ZOOM_MIN) return
  visiblePlaces.value.forEach(addMarkerForPlace)
}

// Debounce
function debounce(fn, ms = 150) {
  let t
  return (...args) => {
    clearTimeout(t)
    t = setTimeout(() => fn(...args), ms)
  }
}

function updateViewState() {
  zoom.value = map.getZoom()
  bounds.value = map.getBounds()
  renderVisibleMarkers()
}
const debouncedUpdate = debounce(updateViewState, 120)

// Local storage persistence
const LS_KEY = 'places.v1'
function persistLocal() {
  try { localStorage.setItem(LS_KEY, JSON.stringify(allPlaces.value)) } catch {}
}
function loadLocal() {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (raw) {
      const arr = JSON.parse(raw)
      if (Array.isArray(arr)) allPlaces.value = arr
    }
  } catch {}
}

function escapeHTML(s) {
  return String(s).replace(/[&<>"']/g, c => ({
    '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'
  }[c]))
}

onMounted(() => {
  map = L.map(mapEl.value, { zoomControl: false }).setView([DEFAULT.lat, DEFAULT.lng], DEFAULT.zoom)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map)

  // Separate layers: keep home pin even when re-rendering place markers
  homeLayer = L.layerGroup().addTo(map)
  markersLayer = L.layerGroup().addTo(map)

  L.control.zoom({ position: 'bottomright' }).addTo(map)

  // click preview (kept)
  map.on('click', (e) => {
    L.popup()
      .setLatLng(e.latlng)
      .setContent(`Lat ${e.latlng.lat.toFixed(5)}, Lng ${e.latlng.lng.toFixed(5)}`)
      .openOn(map)
  })

  map.on('moveend', debouncedUpdate)
  map.on('zoomend', debouncedUpdate)

  loadLocal()
  updateViewState()

  // center on the user's city & country and add a home marker
  centerOnUserBase()
})

onBeforeUnmount(() => {
  if (map) map.remove()
})
</script>

<style scoped>
/* 2-column layout: Map | Sidebar */
.layout {
  display: grid;
  grid-template-columns: 1fr 360px;
  height: calc(100vh - 56px);
  position: relative;
}

/* Map fills left column */
#map {
  position: relative;
  border-top: 1px solid rgba(255,255,255,0.05);
}

/* Sidebar always visible */
.panel {
  position: relative;
  z-index: 2000;
  background: rgba(10,12,20,.9);
  border-left: 1px solid rgba(255,255,255,.08);
  padding: 14px;
  color: #e9ecf1;
  overflow: auto;
}
.panel header { display: flex; align-items: baseline; gap: 10px; justify-content: space-between; margin-bottom: 8px; }
.small { font-size: 12px; }
.card { background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.08); padding: 10px; border-radius: 10px; }

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

/* List */
.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 8px; }
.list li {
  display: grid; grid-template-columns: 12px 1fr; gap: 10px; align-items: center;
  padding: 8px; border-radius: 10px; cursor: pointer; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06);
}
.dot { width: 8px; height: 8px; border-radius: 50%; background: #8bd0ff; }
.txt .t { font-weight: 700; }
.txt .s { font-size: 12px; opacity: .65; }
.muted { opacity: .7; }

/* Tools pinned at bottom of sidebar */
.toolbar {
  position: sticky;
  bottom: 0;
  display: flex; gap: 8px; justify-content: space-between; padding-top: 10px;
}
.toolbar button {
  flex: 1;
  height: 40px; border-radius: 10px; border: 1px solid rgba(255,255,255,.15);
  background: rgba(0,0,0,.35); color: #fff; font-size: 18px; cursor: pointer;
  box-shadow: 0 6px 24px rgba(0,0,0,.25);
}
</style>

<style>
/* Global (unscoped) — applies to Leaflet’s injected HTML */

/* Red pin for saved places */
.pin-red { position: relative; }
.pin-red .pin-head {
  width: 14px; height: 14px; border-radius: 50%;
  background: #ff3b3b; border: 2px solid #fff; box-shadow: 0 0 0 2px rgba(255,59,59,.35);
  transform: translate(-50%, -50%);
  position: absolute; left: 50%; top: 50%;
}
.pin-red .pin-stick {
  width: 2px; height: 8px; background: #ff3b3b;
  position: absolute; left: 50%; bottom: -6px; transform: translateX(-50%);
}

/* Home pin (gold) for user base */
.pin-home { position: relative; }
.pin-home .pin-head {
  width: 16px; height: 16px; border-radius: 50%;
  background: #ffd36a; border: 2px solid #fff; box-shadow: 0 0 0 2px rgba(255,211,106,.35);
  transform: translate(-50%, -50%);
  position: absolute; left: 50%; top: 50%;
}
.pin-home .pin-stick {
  width: 2px; height: 10px; background: #cc9a2e;
  position: absolute; left: 50%; bottom: -8px; transform: translateX(-50%);
}
</style>
