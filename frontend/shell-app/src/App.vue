<script setup>
import { ref, onMounted } from 'vue'
import io from 'socket.io-client'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// --- STATE ---
const alerts = ref([])
const truckStatus = ref({ truck_id: 'Waiting...', speed: 0, fuel: 0 })
const connectionStatus = ref('🔴 Disconnected')
let map = null
let truckMarker = null

onMounted(() => {
  // 1. Initialize the Leaflet Map
  // Centers on Bucharest by default
  map = L.map('map').setView([44.4268, 26.1025], 13)

  // Load OpenStreetMap tiles (free)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap contributors'
  }).addTo(map)

  // Create a custom icon for the truck
  const truckIcon = L.icon({
    iconUrl: 'https://cdn-icons-png.flaticon.com/512/741/741407.png',
    iconSize: [40, 40],
    iconAnchor: [20, 20]
  })

  // 2. Connect to our Node.js WebSocket Service
  const socket = io('http://localhost:3000')

  socket.on('connect', () => {
    connectionStatus.value = '🟢 Connected to Satellite'
  })

  // 3. LISTEN FOR TRUCK UPDATES (From Kafka -> Node -> Here)
  socket.on('truck_update', (data) => {
    truckStatus.value = data
    const { lat, lon } = data.location

    // Update or Create Marker
    if (!truckMarker) {
      truckMarker = L.marker([lat, lon], { icon: truckIcon }).addTo(map)
      map.panTo([lat, lon]) // Follow the truck
    } else {
      const newLatLng = new L.LatLng(lat, lon);
      truckMarker.setLatLng(newLatLng)
      map.panTo(newLatLng)
    }

    // Check for Speeding (Frontend Alert)
    if (data.speed > 100) {
       addLog(`⚠️ High Speed Alert: ${data.speed} km/h`)
    }
  })
})

const addLog = (msg) => {
  const time = new Date().toLocaleTimeString()
  alerts.value.unshift({ id: Date.now(), text: `[${time}] ${msg}` })
  if (alerts.value.length > 7) alerts.value.pop()
}
</script>

<template>
  <div class="dashboard-container">
    <header class="navbar">
      <div class="brand">
        <h1>🚚 FleetCommander <span class="version">v1.0</span></h1>
      </div>
      <div class="status-pill">{{ connectionStatus }}</div>
    </header>

    <div class="main-content">
      <div class="map-panel">
        <div id="map"></div>
      </div>

      <aside class="sidebar">
        <div class="card stat-card">
          <h3>Current Telemetry</h3>
          <div class="stat-row">
            <span class="label">TRUCK ID</span>
            <span class="value small">{{ truckStatus.truck_id }}</span>
          </div>
          <div class="stat-row">
            <span class="label">SPEED</span>
            <span class="value" :class="{ danger: truckStatus.speed > 100 }">
              {{ truckStatus.speed }} <small>km/h</small>
            </span>
          </div>
          <div class="stat-row">
            <span class="label">FUEL</span>
            <span class="value">{{ truckStatus.fuel }}%</span>
          </div>
        </div>

        <div class="card log-card">
          <h3>🚨 Event Log</h3>
          <ul>
            <li v-for="alert in alerts" :key="alert.id">
              {{ alert.text }}
            </li>
          </ul>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
/* Layout */
.dashboard-container { height: 100vh; display: flex; flex-direction: column; background: #111; color: #eee; font-family: 'Segoe UI', sans-serif; }
.main-content { flex: 1; display: flex; overflow: hidden; }

/* Header */
.navbar { background: #1a1a1a; padding: 0 20px; height: 60px; display: flex; align-items: center; justify-content: space-between; border-bottom: 2px solid #00d2ff; }
h1 { font-size: 1.2rem; margin: 0; }
.version { background: #333; font-size: 0.8rem; padding: 2px 6px; border-radius: 4px; margin-left: 10px; }
.status-pill { font-size: 0.9rem; font-weight: bold; }

/* Map */
.map-panel { flex: 2; position: relative; }
#map { width: 100%; height: 100%; }

/* Sidebar */
.sidebar { flex: 1; max-width: 400px; background: #181818; border-left: 1px solid #333; padding: 20px; display: flex; flex-direction: column; gap: 20px; }

/* Cards */
.card { background: #222; border: 1px solid #333; padding: 15px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.3); }
h3 { margin-top: 0; border-bottom: 1px solid #444; padding-bottom: 10px; color: #00d2ff; font-size: 1rem; }

.stat-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.label { color: #888; font-size: 0.8rem; font-weight: bold; }
.value { font-size: 1.5rem; font-weight: bold; font-family: monospace; }
.value.small { font-size: 0.9rem; }
.danger { color: #ff3333; animation: blink 0.5s infinite alternate; }

/* Logs */
ul { list-style: none; padding: 0; margin: 0; }
li { font-size: 0.85rem; padding: 8px 0; border-bottom: 1px solid #333; color: #ffcc00; }

@keyframes blink { from { opacity: 1; } to { opacity: 0.5; } }
</style>