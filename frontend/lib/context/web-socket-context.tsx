
'use client'

import React, { createContext, useContext, useEffect, useState, useRef } from 'react';
import { WebSocketMessage } from '@/types'

interface WebSocketContextType {
  isConnected: boolean;
  lastMessage: WebSocketMessage | null;
  sendMessage: (message: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export const WebSocketProvider: React.FC<{children: React.ReactNode}> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connectWebSocket = () => {
    if (socketRef.current) {
      socketRef.current.close();
    }

    const ws = new WebSocket(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws');
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setLastMessage(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message', error);
        
        console.log('Raw received data:', event.data);
        
        try {
          const jsonPattern = /({[\s\S]*}|\[[\s\S]*\])/;
          const match = event.data.match(jsonPattern);
          if (match && match[0]) {
            const extractedData = JSON.parse(match[0]);
            console.log('Successfully parsed extracted JSON');
            setLastMessage(extractedData);
          }
        } catch (secondError) {
          console.error('Failed to extract valid JSON', secondError);
        }
      }
    };
    
    ws.onclose = (event) => {
      console.log('WebSocket disconnected', event.code, event.reason);
      setIsConnected(false);
      
      if (!document.hidden) {
        scheduleReconnect();
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    socketRef.current = ws;
  };

  const scheduleReconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    reconnectTimeoutRef.current = setTimeout(() => {
      console.log('Attempting to reconnect...');
      connectWebSocket();
    }, 3000);
  };

  const handleVisibilityChange = () => {
    if (document.hidden) {
      // }
    } else {
      // Tab is now visible again
      console.log('Tab visible, checking connection');
      
      // Check if we need to reconnect
      if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
        console.log('Connection lost while tab was inactive, reconnecting');
        connectWebSocket();
      }
    }
  };

  useEffect(() => {
    // Initial connection
    connectWebSocket();

    // Set up visibility change listener
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    // Cleanup on unmount
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      
      if (socketRef.current) {
        socketRef.current.close(1000, 'Component unmounting');
      }
    };
  }, []);
  
  useEffect(() => {
    const interval = setInterval(() => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        if (!document.hidden) {
          socketRef.current.send(JSON.stringify({ type: 'ping' }));
        }
      }
    }, 30000); 
    
    return () => clearInterval(interval);
  }, []);
  
  // Function to send messages
  const sendMessage = (message: any) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(message));
    } else {
      console.warn('Cannot send message, socket is not connected');
    }
  };
  
  return (
    <WebSocketContext.Provider value={{ isConnected, lastMessage, sendMessage }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};