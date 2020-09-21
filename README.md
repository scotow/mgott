# mgott

An HTTP to MQTT publish proxy.

Send a POST request on `/publish` using the following body as a template:

```json
{
  "host": "mqtt.example.com",
  "port": 1883,
  "username": "username",
  "password": "password",
  "topic": "topic",
  "payload": "Hello, World!"
}
```

- `host` defaults to `localhost`;
- `port` defaults to 1883;
- if one of `username` or `password` is missing, none are specified;
- `topic` is mandatory and results in a 400 error if not specified;
- `payload` defaults to an empty string.
