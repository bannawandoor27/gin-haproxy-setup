import asyncio
import json

import requests
import websockets
from fastapi import FastAPI

app = FastAPI()

# The URL of the WebSocket in your Gin application
WS_URL = "ws://34.93.234.147/ws"
WS_URL = "ws://localhost:8080/ws"


async def connect_to_gin_ws():
    """
    Connects to the Gin WebSocket server, handles incoming messages,
    and sends a response object with default values back.
    """
    while True:  # Keep trying to reconnect indefinitely
        try:
            async with websockets.connect(WS_URL) as websocket:
                print("Connected to Gin WebSocket server")
                while True:
                    # Receive a message
                    message = await websocket.recv()
                    request_obj = json.loads(message)
                    # print(f"Received request: {request_obj}")

                    # Prepare a response with default values
                    response_obj = {
                        "status": "success",
                        "data": {
                            "message": "Processed by FastAPI",
                            "defaultKey": 800*5000
                        }
                    }

                    # Send the response back through the WebSocket connection
                    await websocket.send(json.dumps(response_obj))
        except websockets.exceptions.ConnectionClosedError as e:
            print(f"WebSocket connection closed with error: {e}. Reconnecting in 5 seconds...")
        except Exception as e:
            print(f"Error: {e}. Reconnecting in 5 seconds...")
        await asyncio.sleep(5)  # Wait for 5 seconds before trying to reconnect

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(connect_to_gin_ws())

@app.get("/process")
async def process_request():
    """
    A placeholder endpoint to simulate processing a request.
    """
    print("Processing request...")
    return {"message": "Processing request..."}
