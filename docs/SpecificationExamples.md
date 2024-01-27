# Specification Examples
This spec outlines and provides examples for the structure of messages expected by the server over a websocket.

In the examples below *Game* refers to an active instance of the game client (e.g., a television or computer), *Web* refers to active instances of the client web application (e.g., player controlled cell phones), and *Server* refers to this server.

Additionally, any malformed or otherwise improper requests result in a *Connection Refused* response.

```json
{
    "message_type": "connection_refused"
}
```

#### Create lobby (Game -> Server)
```json
{
    "message_type": "create_lobby"
}
```

#### Lobby Join Attempt (Web -> Server)
```json
{
    "message_type": "lobby_join_attempt",
    "lobby_code": <LOBBY_CODE>,
    "player_name": <PLAYER_NAME>
}
```


