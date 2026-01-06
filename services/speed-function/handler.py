def handle(event):
    """
    Pure business logic.
    Input: A dictionary containing telemetry data.
    Output: An alert dictionary if speeding, or None.
    """
    speed = event.get("speed", 0)
    truck_id = event.get("truck_id", "UNKNOWN")
    
    limit = 100 # km/h

    if speed > limit:
        return {
            "alert_type": "SPEEDING_VIOLATION",
            "truck_id": truck_id,
            "speed": speed,
            "limit": limit,
            "fine": (speed - limit) * 10, # 10 Euro per km/h over limit
            "timestamp": "Now"
        }
    
    return None