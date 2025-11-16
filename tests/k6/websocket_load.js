/**
 * WebSocket Performance Test
 * 
 * Tests WebSocket endpoints for streaming logs and real-time data.
 * Measures connection establishment, message latency, and throughput.
 * 
 * Usage:
 *   k6 run tests/k6/websocket_load.js
 */

import ws from 'k6/ws';
import { check, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const wsErrorRate = new Rate('ws_errors');
const wsMessageLatency = new Trend('ws_message_latency');
const wsMessagesReceived = new Counter('ws_messages_received');
const wsConnectionTime = new Trend('ws_connection_time');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const LOG_ENDPOINT = '/api/logs/stream?level=debug';

export const options = {
    stages: [
        { duration: '10s', target: 5 },    // Ramp up to 5 VUs
        { duration: '30s', target: 10 },   // Stay at 10 VUs
        { duration: '10s', target: 0 },    // Ramp down
    ],
    thresholds: {
        'ws_errors': ['rate<0.1'],                   // Less than 10% errors
        'ws_message_latency': ['p(95)<500'],         // 95th percentile < 500ms
        'ws_connection_time': ['p(95)<1000'],        // Connection < 1s
    },
};

export default function () {
    group('WebSocket Log Streaming', () => {
        const wsUrl = `ws://localhost:3000${LOG_ENDPOINT}`;
        
        const startTime = Date.now();
        let connectionEstablished = false;
        let messageCount = 0;
        const messageTimes = [];

        const res = ws.connect(wsUrl, {}, function (socket) {
            // Connection established
            const connectionTime = Date.now() - startTime;
            wsConnectionTime.add(connectionTime);
            connectionEstablished = true;

            socket.on('open', () => {
                check(socket.readyState, {
                    'ws connection open': (r) => r === ws.OPEN,
                });
            });

            socket.on('message', (message) => {
                const receiveTime = Date.now();
                messageCount++;
                wsMessagesReceived.add(1);

                try {
                    const data = JSON.parse(message);
                    
                    check(data, {
                        'message has level': (d) => d.level !== undefined,
                        'message has timestamp': (d) => d.timestamp !== undefined,
                        'message has content': (d) => d.content !== undefined,
                    });

                    // Calculate message latency (server processing time if included)
                    if (data.server_time) {
                        const latency = receiveTime - data.server_time;
                        wsMessageLatency.add(latency);
                        messageTimes.push(latency);
                    }
                } catch (e) {
                    wsErrorRate.add(1);
                    check(false, {
                        'message is valid JSON': () => false,
                    });
                }

                // Close after receiving 50 messages or 30 seconds
                if (messageCount >= 50) {
                    socket.close();
                }
            });

            socket.on('close', () => {
                check(socket.readyState, {
                    'ws connection closed': (r) => r === ws.CLOSED,
                });

                const success = connectionEstablished && messageCount > 10;
                check(success, {
                    'received at least 10 messages': () => messageCount >= 10,
                    'connection established': () => connectionEstablished,
                });

                if (!success) {
                    wsErrorRate.add(1);
                }
            });

            socket.on('error', (e) => {
                wsErrorRate.add(1);
                check(false, {
                    'no ws errors': () => false,
                });
            });

            // Close after 20 seconds if still open
            socket.setTimeout(() => {
                socket.close();
            }, 20000);
        });

        check(res, {
            'ws response status is 101': (r) => r.status === 101,
        });
    });
}

export function handleSummary(data) {
    console.log('=== WebSocket Performance Test Summary ===');
    console.log(`Messages Received: ${data.metrics.ws_messages_received.values.count}`);
    console.log(`WS Errors: ${data.metrics.ws_errors.values.count}`);
    console.log(`Error Rate: ${(data.metrics.ws_errors.values.rate * 100).toFixed(2)}%`);
    console.log(`Avg Connection Time: ${data.metrics.ws_connection_time.values.avg.toFixed(0)}ms`);
    console.log(`Avg Message Latency: ${data.metrics.ws_message_latency.values.avg.toFixed(0)}ms`);
    console.log(`P95 Message Latency: ${data.metrics.ws_message_latency.values['p(95)'].toFixed(0)}ms`);
    
    return {
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
    };
}
