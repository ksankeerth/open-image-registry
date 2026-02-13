import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import AppRouter from './AppRouter';
import 'primeflex/primeflex.css';
import 'primeicons/primeicons.css';
import './assets/themes/viva-light/theme.css';
import './assets/themes/viva-light/custom.css';
import './index.css';
import { ToastProvider } from './components/ToastComponent';
import { PrimeReactProvider } from 'primereact/api';

import '@fontsource/montserrat';
import '@fontsource/montserrat/500.css';
import '@fontsource/montserrat/700.css';
import '@fontsource/inter';
import { client } from './api/client.gen';
import { LoaderProvider } from './components/loader';

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement);

client.setConfig({
  baseUrl: window.APP_CONFIG?.API_BASE_URL || 'http://localhost:8000/api/v1',
  parseAs: 'json',
});

client.interceptors.response.use(async (response) => {
  if (response.status === 401) {
    setTimeout(() => {
      window.location.href = '/login';
    }, 100);
  }

  return response;
});

client.interceptors.response.use(async (response) => {
  if (response.status === 401) {
    setTimeout(() => {
      window.location.href = '/login';
    }, 100);
  }

  return response;
});

root.render(
  <React.StrictMode>
    <ToastProvider>
      <LoaderProvider>
        <PrimeReactProvider
          value={{
            hideOverlaysOnDocumentScrolling: true,
          }}
        >
          <RouterProvider router={AppRouter} />
        </PrimeReactProvider>
      </LoaderProvider>
    </ToastProvider>
  </React.StrictMode>
);
