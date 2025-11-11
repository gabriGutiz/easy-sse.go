# SSE Broadcaster Simple Example

This example demonstrates a basic implementation of **Server-Sent Events (SSE)** using a simple HTML/JavaScript, with a Python script acting as a publisher/broadcaster.

## Structure

  * `index.html`: A simple HTML page with JavaScript that connects to the SSE channel.
  * `broadcast.py`: A Python script to send messages (broadcast) to a specific channel on the server.

## Getting Started

### 1. Run the Go Server

The server will listen on `http://127.0.0.1:8080`.

```bash
go run main.go
```

### 2. Open the Subscriber (Client)

Open the `index.html` file in your web browser.

The JavaScript in `index.html` immediately attempts to connect to the SSE channel named **`foo`**:

```javascript
const eventSrc = new EventSource("http://127.0.0.1:8080/channels/foo");
// ...
```

You should see a log message in the server console confirming the connection:

```
2025/11/14 12:40:18 listener connected: id=foo conn=...
```

### 3. Broadcast a Message (Publisher)

Run the Python script to send **ten sequential messages** to the channel **`foo`**.
The script sends a `POST` request to the `/channels/{id}/broadcast` endpoint.

```bash
./broadcast.py
```

### 4. Observe the Results

  * **Server Console:** You'll see **ten** logs confirming the messages were sent:

    ```
    2025/11/14 12:40:18 sent to foo: foo=bar
    2025/11/14 12:40:18 sent to foo: foo=bar
    ... (8 more logs)
    ```

  * **Browser Window:** **Ten** new items will be appended to the list:

      * **message: foo=bar**
      * **message: foo=bar**
      * **... (8 more list items)**

## Explanation

### 1. HTML Subscriber (`index.html`)

  * It uses the **`EventSource`** API, which is the standard browser mechanism for consuming SSE.
  * `new EventSource("http://127.0.0.1:8080/channels/foo")` establishes the connection.
  * The `eventSrc.onmessage` handler is triggered whenever the server sends a message using the `data: ...\n\n` format. The message content is available as `event.data`.

### 2. Python Publisher (`broadcast.py`)

  * Uses the standard **`requests`** library to make **POST** calls to the broadcast endpoint.
  * The script now includes a **`for _ in range(10):`** loop to send the message **ten times** in quick succession.
  * The data sent in the request body (`{"foo": "bar"}`) is what the Go server will read and relay to the subscribers.
  * `response.raise_for_status()` ensures the script stops if any of the requests fail.
  * **Note**: The Go server is non-blocking. If the connections are slow, you might see "channel busy" logs in the Go console, indicating some messages were dropped.
