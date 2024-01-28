# Specification Examples
This spec outlines and provides examples for the structure of messages expected by the server over a websocket.

In the examples below *Game* refers to an active instance of the game client (e.g., a television or computer), *Web* refers to active instances of the client web application (e.g., player controlled cell phones), and *Server* refers to this server.

Additionally, any malformed or otherwise improper requests result in a *Connection Refused* response.

```json
{
    "message_type": "connection_refused"
}
```
### Create lobby (Game -> Server)
#### Request
```json
{
    "message_type": "create_lobby"
}
```

#### Response (Game)
```json
{ 
    "message_type": "lobby_code", 
    "lobby_code": "1234" 
}
```

### Lobby Join Attempt (Web -> Server)
#### Request
```json
{
    "message_type": "lobby_join_attempt",
    "lobby_code": "<LOBBY_CODE>",
    "player_name": "<PLAYER_NAME>"
}
```

#### Response (Web)
```json
{
    "message_type": "player_id",
    "player_id": "<PLAYER_UUID>"
}
```

#### Response (Game)
```json
{
    "message_type": "player_joined",
    "player": {
        "player_id": "<PLAYER_UUID>",
        "name": "<PLAYER_NAME>"
    }
}
```

### Game Start (Game -> Server)
#### Request
```json
{
    "message_type": "game_start",
}
```

#### Response (Server -> Web & Server -> Game)
```json
{
    "message_type": "game_start",
    "number_of_jobs": "<NUMBER_OF_JOBS_REQUIRED_PER_PLAYER>"
}
```

### Job Submitted (Web -> Server)
#### Request (Consumed by the server, no immediate response)
```json
{
    "message_type": "job_submitted",
    "job_input": "<JOB_INPUT>"
}
```

### Player Job Submitting Finished (Server -> Game)
#### Response (Sent when individual web clients submit the required number of jobs)
```json
{
    "message_type": "player_job_submitting_finished",
    "player_id": "<PLAYER_UUID>"
}
```

### Received Cards (Server -> Web / Server -> Game)
#### Response (Server -> Web)
```json
{
    "message_type": "received_cards",
    "drawn_cards": {
        "card_id": "<CARD_ID>",
        "job_text": "<JOB_CARD_TEXT>"
    },
    "job_card": "<USER_JOB_CARD>"
}
```

#### Response (Server -> Game)
```json
{
    "message_type": "received_cards"
}
```
