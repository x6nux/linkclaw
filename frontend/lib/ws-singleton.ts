import { WSClient } from "./ws-client";

let _client: WSClient | null = null;

export function getWSClient(): WSClient {
  if (!_client) {
    const wsBase =
      typeof window !== "undefined"
        ? (process.env.NEXT_PUBLIC_WS_URL ?? `ws://${window.location.host}/api/v1/messages/ws`)
        : "ws://localhost/api/v1/messages/ws";
    _client = new WSClient(wsBase);
    _client.connect();
  }
  return _client;
}

export function destroyWSClient() {
  _client?.disconnect();
  _client = null;
}
