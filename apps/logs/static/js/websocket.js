// LogsWebSocket - Handles WebSocket connection for real-time log streaming
class LogsWebSocket {
  constructor(url, onMessage, onStatusChange) {
    this.url = url;
    this.onMessage = onMessage;
    this.onStatusChange = onStatusChange;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.isPaused = false;
    // Enable debug logging only in development mode
    this.debugEnabled = window.location.hostname === 'localhost' || 
                        window.location.hostname === '127.0.0.1' ||
                        window.DEBUG_ENABLED === true;
  }

  // Internal debug logger - only logs if debugEnabled
  _debug(message, ...args) {
    if (this.debugEnabled) {
      console.log(`[WebSocket] ${message}`, ...args);
    }
  }

  _error(message, ...args) {
    if (this.debugEnabled) {
      console.error(`[WebSocket] ${message}`, ...args);
    }
  }

  connect() {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this._debug('Connected');
        this.reconnectAttempts = 0;
        this.onStatusChange('connected');
      };

      this.ws.onmessage = (event) => {
        if (!this.isPaused) {
          try {
            const logEntry = JSON.parse(event.data);
            this.onMessage(logEntry);
          } catch (e) {
            this._error('Failed to parse log entry:', e);
          }
        }
      };

      this.ws.onerror = (error) => {
        this._error('Error:', error);
        this.onStatusChange('error');
      };

      this.ws.onclose = () => {
        this._debug('Closed');
        this.onStatusChange('disconnected');
        this.attemptReconnect();
      };
    } catch (error) {
      this._error('Failed to create WebSocket:', error);
      this.onStatusChange('error');
      this.attemptReconnect();
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      this._debug(`Reconnecting... (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      this.onStatusChange('reconnecting');

      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      this._error('Max reconnect attempts reached');
      this.onStatusChange('failed');
    }
  }

  pause() {
    this.isPaused = true;
  }

  resume() {
    this.isPaused = false;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
