
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

    // Get WebSocket URL from environment or determine based on current protocol
    let wsUrl = process.env.NEXT_PUBLIC_WS_URL;
    
    if (!wsUrl) {
      // Auto-detect protocol based on current page protocol
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      wsUrl = `${protocol}//${host}/ws`;
    }
    
    console.log('Connecting to WebSocket URL:', wsUrl);
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connection established successfully');
      setIsConnected(true);
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setLastMessage(data);
      } catch (error) {
        
        try {
          const jsonPattern = /({[\s\S]*}|\[[\s\S]*\])/;
          const match = event.data.match(jsonPattern);
          if (match && match[0]) {
            const extractedData = JSON.parse(match[0]);
            setLastMessage(extractedData);
          }
        } catch (secondError) {
        }
      }
    };
    
    ws.onclose = (event) => {
      console.log('WebSocket connection closed', event.code, event.reason);
      setIsConnected(false);
      
      if (!document.hidden) {
        scheduleReconnect();
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error occurred:', error);
    };
    
    socketRef.current = ws;
  };

  const scheduleReconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    reconnectTimeoutRef.current = setTimeout(() => {
      connectWebSocket();
    }, 3000);
  };

  const handleVisibilityChange = () => {
    if (document.hidden) {
      // }
    } else {
      
      // Check if we need to reconnect
      if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
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