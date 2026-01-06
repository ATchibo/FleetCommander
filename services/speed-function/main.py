import asyncio
import json
from aiokafka import AIOKafkaConsumer, AIOKafkaProducer
from handler import handle  # Import your function

# Config
KAFKA_BROKER = "localhost:9092"
SOURCE_TOPIC = "vehicle-telemetry"
DEST_TOPIC = "system-alerts"

async def start_function_trigger():
    # 1. Consumer (The Trigger)
    consumer = AIOKafkaConsumer(
        SOURCE_TOPIC,
        bootstrap_servers=KAFKA_BROKER,
        group_id="faas-speed-group"
    )
    
    # 2. Producer (To send the output)
    producer = AIOKafkaProducer(bootstrap_servers=KAFKA_BROKER)

    await consumer.start()
    await producer.start()
    print("⚡ Speed Function Trigger Armed. Watching for > 100km/h...")

    try:
        async for msg in consumer:
            # Decode Event
            event_data = json.loads(msg.value.decode("utf-8"))
            
            # --- INVOKE FUNCTION ---
            result = handle(event_data)
            # -----------------------

            if result:
                # If function returned data, publish it to 'system-alerts'
                print(f"🚨 VIOLATION DETECTED: Truck {result['truck_id']} doing {result['speed']}km/h")
                
                await producer.send_and_wait(
                    DEST_TOPIC, 
                    json.dumps(result).encode("utf-8")
                )
    finally:
        await consumer.stop()
        await producer.stop()

if __name__ == "__main__":
    asyncio.run(start_function_trigger())