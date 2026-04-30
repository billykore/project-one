const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";

export const api = {
  post: async <T>(endpoint: string, data: unknown): Promise<T> => {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  },
  
  // Add other methods (get, put, delete) as needed in the future
};
