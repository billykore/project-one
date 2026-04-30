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
      let errorMessage = `HTTP error! status: ${response.status}`;
      const contentType = response.headers.get("content-type");
      
      if (contentType && contentType.includes("application/json")) {
        const errorData = await response.json().catch(() => ({}));
        if (errorData.error) {
          errorMessage = errorData.error;
        }
      }
      
      throw new Error(errorMessage);
    }

    return response.json();
  },
  
  // Add other methods (get, put, delete) as needed in the future
};
