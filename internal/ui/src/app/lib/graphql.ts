// Simple fetch-based GraphQL client

export interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{
    message: string;
    locations?: Array<{
      line: number;
      column: number;
    }>;
    path?: string[];
  }>;
}

export async function graphqlRequest<T>(
  query: string,
  variables?: Record<string, unknown>,
  operationName?: string
): Promise<GraphQLResponse<T>> {
  const response = await fetch('/graph/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query,
      variables,
      operationName,
    }),
    credentials: 'include', // Ensures cookies are sent with the request
  });

  return response.json();
}
