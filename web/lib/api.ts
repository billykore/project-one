const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";
const JSON_HEADERS = { "Content-Type": "application/json" };

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
    let errorMessage = `Something went wrong! (${response.status})`;
    const contentType = response.headers.get("content-type");

    if (contentType && contentType.includes("application/json")) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.error) errorMessage = errorData.error;
    }

    throw new ApiError(errorMessage, response.status);
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return response.json();
  }

  return {} as T;
};

async function request<T>(method: "GET" | "POST" | "PUT" | "DELETE", endpoint: string, data?: unknown): Promise<T> {
  const response = await fetch(`${BASE_URL}${endpoint}`, {
    method,
    headers: JSON_HEADERS,
    body: data === undefined ? undefined : JSON.stringify(data),
    credentials: "include",
  });

  return handleResponse<T>(response);
}

export const api = {
  get: async <T>(endpoint: string): Promise<T> => request<T>("GET", endpoint),

  post: async <T>(endpoint: string, data: unknown): Promise<T> => request<T>("POST", endpoint, data),

  put: async <T>(endpoint: string, data: unknown): Promise<T> => request<T>("PUT", endpoint, data),

  delete: async <T>(endpoint: string): Promise<T> => request<T>("DELETE", endpoint),
};
