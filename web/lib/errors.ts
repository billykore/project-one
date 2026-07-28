export interface ValidationError {
  field: string;
  reason: string;
  message: string;
}

export interface ProblemDetail {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance: string;
  code?: string;
  request_id?: string;
  errors?: ValidationError[];
}

export class ApiError extends Error {
  status: number;
  code?: string;
  type?: string;
  instance?: string;
  validationErrors?: ValidationError[];

  constructor(problem: ProblemDetail) {
    super(problem.detail || problem.title);
    this.status = problem.status;
    this.code = problem.code;
    this.type = problem.type;
    this.instance = problem.instance;
    this.validationErrors = problem.errors;
    this.name = "ApiError";
  }
}

export async function handleApiResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    // Global 401 handling: redirect to login, preserving the current path for post-login redirect.
    if (response.status === 401 && typeof window !== "undefined") {
      const currentPath = window.location.pathname;
      if (currentPath !== "/login") {
        window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
      }
      throw new ApiError({
        type: "about:blank",
        title: "Unauthorized",
        status: 401,
        detail: "Session expired. Redirecting to login...",
        instance: currentPath,
      });
    }

    let problem: ProblemDetail = {
      type: "about:blank",
      title: "Unknown Error",
      status: response.status,
      detail: `Something went wrong! (${response.status})`,
      instance: "",
    };

    const contentType = response.headers.get("content-type");
    if (contentType && (contentType.includes("application/problem+json") || contentType.includes("application/json"))) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.type) {
        problem = errorData as ProblemDetail;
      }
    }

    throw new ApiError(problem);
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return response.json();
  }

  return {} as T;
}
