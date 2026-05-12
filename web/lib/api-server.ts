import { cookies } from "next/headers";
import { ApiError } from "./api";

const BASE_URL = process.env.API_URL || "http://localhost:8080";

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    let errorMessage = `Something went wrong! (${response.status})`;
    const contentType = response.headers.get("content-type");
    
    if (contentType && contentType.includes("application/json")) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.error) {
        errorMessage = errorData.error;
      }
    }
    
    throw new ApiError(errorMessage, response.status);
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return response.json();
  }
  
  return {} as T;
};

export const apiServer = {
  get: async <T>(endpoint: string): Promise<T> => {
    const cookieStore = await cookies();
    const cookieString = cookieStore.toString();

    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Cookie": cookieString,
      },
    });

    return handleResponse<T>(response);
  },

  post: async <T>(endpoint: string, data: unknown): Promise<T> => {
    const cookieStore = await cookies();
    const cookieString = cookieStore.toString();

    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Cookie": cookieString,
      },
      body: JSON.stringify(data),
    });

    return handleResponse<T>(response);
  },
};
