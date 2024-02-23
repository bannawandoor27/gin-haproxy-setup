# fastapi_client.py
import asyncio
import json
import websockets
from fastapi import FastAPI

app = FastAPI()

# The URL of the WebSocket in your Gin application
WS_URL = "ws://localhost:8080/ws"

async def connect_to_gin_ws():
    """
    Connects to the Gin WebSocket server and prints messages received from it.
    """
    async with websockets.connect(WS_URL) as websocket:
        # Wait for messages from the Gin server and print them
        while True:
            message = await websocket.recv()
            print(f"Message from Gin server: {message}")

@app.on_event("startup")
async def startup_event():
    """Modified to print the deserialized request object."""
    async def connect_to_gin_ws():
        async with websockets.connect(WS_URL) as websocket:
            while True:
                message = await websocket.recv()
                request_obj = json.loads(message)  # Deserialize the JSON back into a Python dict
                print(f"Received request: {request_obj}")
                # Now you can access request_obj["method"], request_obj["path"], etc.

    asyncio.create_task(connect_to_gin_ws())

@app.get("/process")
async def process_request():
    """
    A placeholder endpoint to simulate processing a request.
    This could be expanded to forward requests to the Gin server via WebSocket.
    """
    print("Processing request...")
    # Here you would add logic to send data to the Gin WebSocket server if needed
    return {"message": "Processing request..."}
