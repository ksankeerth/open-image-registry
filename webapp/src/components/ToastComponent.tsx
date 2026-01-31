import React, { createContext, ReactNode, useContext, useState, useCallback } from 'react';

// Types
interface ToastMessage {
  id: string;
  severity: 'success' | 'warn' | 'error' | 'info';
  detail: string;
  life: number;
  timestamp: number;
}

interface ToastContextType {
  showSuccess: (message: string, life?: number) => void;
  showWarning: (message: string, life?: number) => void;
  showError: (message: string, life?: number) => void;
  showInfo: (message: string, life?: number) => void;
  clear: () => void;
}

interface ToastProviderProps {
  children: ReactNode;
}

// Context
const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const useToast = (): ToastContextType => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

// Individual Toast Component
const ToastItem: React.FC<{
  message: ToastMessage;
  onRemove: (id: string) => void;
}> = ({ message, onRemove }) => {
  const getSeverityConfig = (severity: string) => {
    switch (severity) {
      case 'success':
        return {
          color: '#58cd9a',
          title: 'Success',
        };
      case 'warn':
      case 'warning':
        return {
          color: '#f0ad4e',
          title: 'Warning',
        };
      case 'error':
        return {
          color: '#dc3545',
          title: 'Failure',
        };
      case 'info':
        return {
          color: '#17a2b8',
          title: 'Info',
        };
      default:
        return {
          color: '#58cd9a',
          title: 'Info',
        };
    }
  };

  const config = getSeverityConfig(message.severity);

  // Auto-remove after specified life time
  React.useEffect(() => {
    if (message.life > 0) {
      const timer = setTimeout(() => {
        onRemove(message.id);
      }, message.life);

      return () => clearTimeout(timer);
    }
  }, [message.id, message.life, onRemove]);

  return (
    <div
      className="toast-item"
      style={{
        animation: 'slideInFromRight 0.4s ease-out',
        marginBottom: '0.5rem',
      }}
    >
      <div
        className="flex justify-content-between w-20rem m-2 p-2 border-round-xl shadow-2 surface-0"
        style={{
          borderStyle: 'solid',
          borderWidth: 0,
          borderLeftWidth: '4px',
          borderColor: config.color,
          zIndex: 1000,
        }}
      >
        <div className="flex flex-column">
          <div className="font-semibold">{config.title}</div>
          <div className="text-xs text-color-secondary mt-1">{message.detail}</div>
        </div>
        <div className="flex align-items-center">
          <i
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                onRemove(message.id);
              }
            }}
            className="pi pi-times cursor-pointer text-color-secondary hover:text-color"
            style={{ fontSize: '0.75rem' }}
            onClick={() => onRemove(message.id)}
          />
        </div>
      </div>
    </div>
  );
};

// Toast Container Component
const ToastContainer: React.FC<{
  messages: ToastMessage[];
  onRemove: (id: string) => void;
}> = ({ messages, onRemove }) => {
  if (messages.length === 0) return null;

  return (
    <div
      className="toast-container"
      style={{
        position: 'fixed',
        top: '1rem',
        right: '1rem',
        zIndex: 9999,
        pointerEvents: 'none',
      }}
    >
      {messages.map((message) => (
        <div key={message.id} style={{ pointerEvents: 'auto' }}>
          <ToastItem message={message} onRemove={onRemove} />
        </div>
      ))}
    </div>
  );
};

// Toast Provider
export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [messages, setMessages] = useState<ToastMessage[]>([]);

  // Generate unique ID for each toast
  const generateId = () => `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

  // Add a new toast message
  const addMessage = useCallback(
    (severity: ToastMessage['severity'], detail: string, life: number) => {
      const newMessage: ToastMessage = {
        id: generateId(),
        severity,
        detail,
        life,
        timestamp: Date.now(),
      };

      setMessages((prev) => [...prev, newMessage]);
    },
    []
  );

  // Remove a toast message
  const removeMessage = useCallback((id: string) => {
    setMessages((prev) => prev.filter((msg) => msg.id !== id));
  }, []);

  // Clear all messages
  const clear = useCallback(() => {
    setMessages([]);
  }, []);

  // Toast methods
  const showSuccess = useCallback(
    (message: string, life: number = 5000) => {
      addMessage('success', message, life);
    },
    [addMessage]
  );

  const showWarning = useCallback(
    (message: string, life: number = 5000) => {
      addMessage('warn', message, life);
    },
    [addMessage]
  );

  const showError = useCallback(
    (message: string, life: number = 6000) => {
      addMessage('error', message, life);
    },
    [addMessage]
  );

  const showInfo = useCallback(
    (message: string, life: number = 5000) => {
      addMessage('info', message, life);
    },
    [addMessage]
  );

  const contextValue: ToastContextType = {
    showSuccess,
    showWarning,
    showError,
    showInfo,
    clear,
  };

  return (
    <ToastContext.Provider value={contextValue}>
      {children}
      <ToastContainer messages={messages} onRemove={removeMessage} />
    </ToastContext.Provider>
  );
};
