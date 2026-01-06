import asyncio
import json
import random
import uuid
from aiokafka import AIOKafkaProducer

# Configuration
KAFKA_TOPIC = "vehicle-telemetry"
KAFKA_BOOTSTRAP_SERVERS = "localhost:9092"

# Simulate a specific truck
TRUCK_ID = f"TRUCK-{str(uuid.uuid4())[:8]}"

def generate_telemetry():
    """Generates fake GPS data moving generally northeast"""
    base_lat = 44.4268  # Starting near Bucharest ;)
    base_lon = 26.1025
    
    # Add random jitter to simulate movement
    lat = base_lat + (random.random() * 0.01)
    lon = base_lon + (random.random() * 0.01)
    speed = random.randint(40, 110) # km/h
    
    return {
        "truck_id": TRUCK_ID,
        "location": {"lat": lat, "lon": lon},
        "speed": speed,
        "fuel": random.randint(10, 100),
        "status": "MOVING"
    }

async def send_telemetry():
    producer = AIOKafkaProducer(bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS)
    
    # Connect to Redpanda
    await producer.start()
    print(f"🚛 Truck {TRUCK_ID} engine started. Connected to Kafka/Redpanda.")

    try:
        while True:
            data = generate_telemetry()
            payload = json.dumps(data).encode("utf-8")
            
            # Send the message
            await producer.send_and_wait(KAFKA_TOPIC, payload)
            print(f"📡 Sent: {data['speed']}km/h | {data['location']}")
            
            # Wait 1 second
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        print("Stopping truck...")
    finally:
        await producer.stop()

if __name__ == "__main__":
    asyncio.run(send_telemetry())