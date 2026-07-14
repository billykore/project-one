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
