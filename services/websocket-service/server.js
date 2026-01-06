const { Kafka } = require('kafkajs');
const express = require('express');
const http = require('http');
const { Server } = require("socket.io");
const cors = require('cors');

// 1. Setup Express & Socket.IO
const app = express();
app.use(cors());
const server = http.createServer(app);
const io = new Server(server, {
    cors: { origin: "*" } // Allow any browser to connect
});

// 2. Setup Kafka Consumer (Redpanda)
const kafka = new Kafka({
    clientId: 'dashboard-service',
    brokers: ['localhost:9092']
});

const consumer = kafka.consumer({ groupId: 'dashboard-group' });

async function run() {
    // Connect to Redpanda
    await consumer.connect();
    console.log("✅ Connected to Redpanda");

    // Subscribe to the telemetry topic
    await consumer.subscribe({ topic: 'vehicle-telemetry', fromBeginning: false });

    // 3. Process each message
    await consumer.run({
        eachMessage: async ({ topic, partition, message }) => {
            const rawData = message.value.toString();
            console.log(`📥 Received: ${rawData}`);

            // 4. FORWARD TO BROWSER via WebSocket
            // We broadcast to everyone connected
            io.emit("truck_update", JSON.parse(rawData));
        },
    });
}

// Handle WebSocket connections (just for logging)
io.on('connection', (socket) => {
    console.log('👤 A user connected to the dashboard');
});

// Start the server
run().catch(console.error);
server.listen(3000, () => {
    console.log('🚀 WebSocket Server listening on port 3000');
});