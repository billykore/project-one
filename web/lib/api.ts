const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    let errorMessage = `HTTP error! status: ${response.status}`;
    const contentType = response.headers.get("content-type");
    
    if (contentType && contentType.includes("application/json")) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.error) {
        errorMessage = errorData.error;
      }
    }
    
    throw new ApiError(errorMessage, response.status);
  }

  // Handle empty responses (like 204 No Content) or success messages
  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return response.json();
  }
  
  return {} as T;
};

export const api = {
  get: async <T>(endpoint: string): Promise<T> => {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    });

    return handleResponse<T>(response);
  },

  post: async <T>(endpoint: string, data: unknown): Promise<T> => {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
      credentials: "include",
    });

    return handleResponse<T>(response);
  },
};
