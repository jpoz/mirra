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

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ApolloProvider client={apolloClient}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </QueryClientProvider>
    </ApolloProvider>
  </React.StrictMode>
);
