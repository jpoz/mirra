import { ApolloClient, InMemoryCache, HttpLink, ApolloLink } from '@apollo/client';

// Create the http link that connects to our GraphQL API
const httpLink = new HttpLink({
  uri: '/graph/query',
  credentials: 'include', // Ensures cookies are sent with the request
});

// Create error handling middleware
const errorLink = new ApolloLink((operation, forward) => {
  return forward(operation).map((response) => {
    // If there are GraphQL errors, you can handle them here
    if (response.errors && response.errors.length > 0) {
      console.error('GraphQL Errors:', response.errors);
    }
    return response;
  });
});

// Initialize Apollo Client
export const apolloClient = new ApolloClient({
  link: ApolloLink.from([errorLink, httpLink]),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          // Add type policies as needed
        }
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
      errorPolicy: 'all',
    },
    query: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all',
    },
    mutate: {
      errorPolicy: 'all',
    },
  },
});