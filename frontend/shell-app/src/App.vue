<script setup>
import { ref, onMounted } from 'vue'
import io from 'socket.io-client'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// --- AUTH & ORDERS STATE ---
const token = ref(null) // Stores the JWT
const loginForm = ref({ username: 'admin', password: 'password123' })
const orderForm = ref({ customer: '', destination: '', price: 100 })
const orderStatus = ref('')

// 1. LOGIN FUNCTION
const handleLogin = async () => {
  try {
    const res = await fetch('http://localhost:3002/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(loginForm.value)
    })
    
    if (!res.ok) throw new Error('Login Failed')
    
    const data = await res.json()
    token.value = data.token
    addLog('🔐 Logged in as Admin')
  } catch (err) {
    addLog('❌ Auth Error: Check credentials')
  }
}

// 2. DISPATCH ORDER FUNCTION
const sendOrder = async () => {
  if (!token.value) return

  try {
    const res = await fetch('http://localhost:3001/orders', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token.value}` // Send the JWT!
      },
      body: JSON.stringify({
        id: Math.floor(Math.random() * 1000).toString(),
        customer: orderForm.value.customer,
        destination: orderForm.value.destination,
        price: parseInt(orderForm.value.price)
      })
    })

    if (!res.ok) throw new Error('Order Failed')

    addLog(`📦 Order Dispatched to ${orderForm.value.destination}`)
    orderStatus.value = 'Sent!'
    setTimeout(() => orderStatus.value = '', 2000)
    
    // Clear form
    orderForm.value.customer = ''
    orderForm.value.destination = ''
  } catch (err) {
    addLog('❌ Order Rejected: ' + err.message)
  }
}

// --- STATE ---
const alerts = ref([])
const connectionStatus = ref('🔴 Disconnected')
const selectedId = ref(null) // Track which truck user clicked

// Default state for the sidebar
const truckStatus = ref({ 
  truck_id: 'Select a Truck...', 
  speed: 0, 
  fuel: 0,
  status: 'Waiting'
})

let map = null
const markers = {} // Dictionary to track all markers

onMounted(() => {
  // 1. Initialize Map
  map = L.map('map').setView([44.4268, 26.1025], 13)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap'
  }).addTo(map)

  // 2. NEW ICON: Heavy Delivery Truck
  const truckIcon = L.icon({
    iconUrl: 'https://cdn-icons-png.flaticon.com/512/2554/2554936.png', // <-- New Image
    iconSize: [46, 46],
    iconAnchor: [23, 23], // Center the icon
    popupAnchor: [0, -20]
  })

  // 3. Connect to WebSocket
  const socket = io('http://localhost:3000')

  socket.on('connect', () => {
    connectionStatus.value = '🟢 Connected'
  })

  socket.on('truck_update', (data) => {
    const { truck_id, location, speed } = data

    // LOGIC: Update sidebar ONLY if this is the currently selected truck
    if (selectedId.value === truck_id) {
       truckStatus.value = data
    }

    // MAP LOGIC
    if (!markers[truck_id]) {
      // Create new marker
      const newMarker = L.marker([location.lat, location.lon], { icon: truckIcon })
        .addTo(map)
        .bindPopup(`<b>${truck_id}</b>`)

      // ADD CLICK EVENT: Select this truck when clicked
      newMarker.on('click', () => {
        selectedId.value = truck_id
        truckStatus.value = data // Update sidebar instantly
        addLog(`Monitoring: ${truck_id.substring(0,8)}`)
        
        // Highlight the marker (optional visual cue)
        newMarker.openPopup()
      })

      // Save to dictionary
      markers[truck_id] = newMarker
      addLog(`New Truck Found: ${truck_id.substring(0,8)}`)
    } else {
      // Move existing marker
      const newLatLng = new L.LatLng(location.lat, location.lon)
      markers[truck_id].setLatLng(newLatLng)
      
      // Update popup content dynamically
      markers[truck_id].setPopupContent(`<b>${truck_id}</b><br>Speed: ${speed} km/h`)
    }

    // Global Alert (shows regardless of selection)
    if (speed > 100) {
       addLog(`⚠️ Speeding: ${truck_id.substring(0,8)} (${speed} km/h)`)
    }
  })
})

const addLog = (msg) => {
  const time = new Date().toLocaleTimeString()
  alerts.value.unshift({ id: Date.now(), text: `[${time}] ${msg}` })
  if (alerts.value.length > 8) alerts.value.pop()
}
</script>

<template>
  <div class="dashboard-container">
    <header class="navbar">
      <div class="brand">
        <h1>🚚 FleetCommander <span class="version">PRO</span></h1>
      </div>
      <div class="status-pill">{{ connectionStatus }}</div>
    </header>

    <div class="main-content">
      <div class="map-panel">
        <div id="map"></div>
        <div class="map-overlay" v-if="!selectedId">
          👇 Click a truck to view live telemetry
        </div>
      </div>

      <aside class="sidebar">
        <div class="card command-card">
          <h3>📦 Dispatch Command</h3>
          
          <div v-if="!token" class="login-box">
            <input v-model="loginForm.username" placeholder="Username" />
            <input v-model="loginForm.password" type="password" placeholder="Password" />
            <button @click="handleLogin">ACCESS TERMINAL</button>
          </div>

          <div v-else class="order-box">
            <div class="input-group">
              <label>Customer</label>
              <input v-model="orderForm.customer" placeholder="e.g. Ikea Industries" />
            </div>
            <div class="input-group">
              <label>Destination</label>
              <input v-model="orderForm.destination" placeholder="e.g. Bucharest North" />
            </div>
            <div class="input-group">
              <label>Price (€)</label>
              <input v-model="orderForm.price" type="number" />
            </div>
            
            <button @click="sendOrder" class="dispatch-btn">
              {{ orderStatus || 'DISPATCH DRIVER' }}
            </button>
          </div>
        </div>

        <div class="card stat-card" :class="{ active: selectedId }">
          <h3>
            {{ selectedId ? 'Live Telemetry' : 'No Selection' }}
          </h3>
          
          <div v-if="selectedId">
            <div class="stat-row">
              <span class="label">ID</span>
              <span class="value small">{{ truckStatus.truck_id.substring(0,12) }}...</span>
            </div>
            <div class="stat-row">
              <span class="label">SPEED</span>
              <span class="value" :class="{ danger: truckStatus.speed > 100 }">
                {{ truckStatus.speed }} <small>km/h</small>
              </span>
            </div>
            <div class="stat-row">
              <span class="label">FUEL</span>
              <div class="progress-bar">
                <div class="fill" :style="{ width: truckStatus.fuel + '%' }"></div>
              </div>
              <span class="value-tiny">{{ truckStatus.fuel }}%</span>
            </div>
            <div class="stat-row">
               <span class="label">LAT/LON</span>
               <span class="value-tiny">
                 {{ truckStatus.location?.lat.toFixed(4) }}, {{ truckStatus.location?.lon.toFixed(4) }}
               </span>
            </div>
          </div>
          
          <div v-else class="placeholder-text">
            Select a vehicle on the map to see its status, fuel, and speed.
          </div>
        </div>

        <div class="card log-card">
          <h3>🚨 System Logs</h3>
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
/* Main Layout */
.dashboard-container { height: 100vh; display: flex; flex-direction: column; background: #121212; color: #e0e0e0; font-family: 'Inter', sans-serif; }
.main-content { flex: 1; display: flex; overflow: hidden; }

/* Navbar */
.navbar { background: #1e1e1e; padding: 0 25px; height: 64px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #333; box-shadow: 0 2px 10px rgba(0,0,0,0.5); }
h1 { font-size: 1.3rem; margin: 0; font-weight: 700; color: #fff; }
.version { background: #007acc; font-size: 0.7rem; padding: 2px 6px; border-radius: 4px; margin-left: 10px; text-transform: uppercase; letter-spacing: 1px; }

/* Map */
.map-panel { flex: 3; position: relative; }
#map { width: 100%; height: 100%; z-index: 1; }
.map-overlay { position: absolute; bottom: 30px; left: 50%; transform: translateX(-50%); background: rgba(0,0,0,0.8); padding: 10px 20px; border-radius: 20px; z-index: 1000; pointer-events: none; font-weight: bold; }

/* Sidebar */
.sidebar { flex: 1; min-width: 320px; max-width: 400px; background: #181818; border-left: 1px solid #333; padding: 25px; display: flex; flex-direction: column; gap: 25px; z-index: 2; box-shadow: -5px 0 15px rgba(0,0,0,0.3); }

/* Cards */
.card { background: #232323; border: 1px solid #333; padding: 20px; border-radius: 12px; }
.card.active { border-color: #007acc; box-shadow: 0 0 15px rgba(0, 122, 204, 0.1); }
h3 { margin-top: 0; border-bottom: 1px solid #444; padding-bottom: 15px; color: #00d2ff; font-size: 1.1rem; letter-spacing: 0.5px; }

/* Stats */
.stat-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
.label { color: #888; font-size: 0.75rem; font-weight: 700; letter-spacing: 1px; }
.value { font-size: 1.8rem; font-weight: 700; font-family: 'Roboto Mono', monospace; }
.value.small { font-size: 1rem; color: #ccc; }
.value-tiny { font-size: 0.9rem; font-family: monospace; color: #aaa; }
.danger { color: #ff4d4d; text-shadow: 0 0 10px rgba(255, 77, 77, 0.4); }

/* Fuel Bar */
.progress-bar { flex: 1; height: 8px; background: #333; border-radius: 4px; margin: 0 15px; overflow: hidden; }
.fill { height: 100%; background: linear-gradient(90deg, #ffcc00, #ff9900); transition: width 0.5s ease; }

.placeholder-text { color: #666; font-style: italic; text-align: center; padding: 20px 0; }

/* Command Card Styles */
.command-card input {
  width: 100%;
  background: #333;
  border: 1px solid #444;
  color: white;
  padding: 8px;
  margin-bottom: 10px;
  border-radius: 4px;
  box-sizing: border-box; /* Fix padding width issues */
}

.command-card button {
  width: 100%;
  background: #007acc;
  color: white;
  border: none;
  padding: 10px;
  font-weight: bold;
  cursor: pointer;
  border-radius: 4px;
}
.command-card button:hover { background: #005f9e; }

.input-group { margin-bottom: 8px; }
.input-group label { display: block; font-size: 0.75rem; color: #888; margin-bottom: 4px; }

.dispatch-btn { background: #28a745 !important; }
.dispatch-btn:hover { background: #218838 !important; }

/* Logs */
ul { list-style: none; padding: 0; margin: 0; }
li { font-size: 0.8rem; padding: 10px 0; border-bottom: 1px solid #333; color: #ccc; display: flex; align-items: center; }
li:first-child { color: #fff; font-weight: bold; }
</style>