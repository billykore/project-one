import { cookies } from "next/headers";
import { ApiError } from "./api";

const BASE_URL = process.env.API_URL || "http://localhost:8080";
const JSON_CONTENT_TYPE = "application/json";

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    let errorMessage = `Something went wrong! (${response.status})`;
    const contentType = response.headers.get("content-type");

    if (contentType && contentType.includes(JSON_CONTENT_TYPE)) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.error) errorMessage = errorData.error;
    }

    throw new ApiError(errorMessage, response.status);
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes(JSON_CONTENT_TYPE)) {
    return response.json();
  }

  return {} as T;
};

async function request<T>(method: "GET" | "POST", endpoint: string, data?: unknown): Promise<T> {
  const cookieStore = await cookies();
  const cookieString = cookieStore.toString();
  const cleanEndpoint = endpoint.replace(/^\/api\/v1/, "");

  const response = await fetch(`${BASE_URL}${cleanEndpoint}`, {
    method,
    headers: {
      "Content-Type": JSON_CONTENT_TYPE,
      Cookie: cookieString,
    },
    body: data === undefined ? undefined : JSON.stringify(data),
  });

  return handleResponse<T>(response);
}

export const apiServer = {
  get: async <T>(endpoint: string): Promise<T> => request<T>("GET", endpoint),

  post: async <T>(endpoint: string, data: unknown): Promise<T> => request<T>("POST", endpoint, data),
};
