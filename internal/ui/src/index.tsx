import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './app/App';
import './style/index.css';
import { BrowserRouter } from "react-router";
import { ApolloProvider } from '@apollo/client';
import { apolloClient } from './app/lib/apolloClient';
import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { ThemeProvider } from "./app/components/theme-provider"

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ApolloProvider client={apolloClient}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </ThemeProvider>
      </QueryClientProvider>
    </ApolloProvider>
  </React.StrictMode>
);
