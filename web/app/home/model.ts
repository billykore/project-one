export interface UserResponse {
  username: string;
  email: string;
  name: string;
}

export interface HomeState {
  user: UserResponse | null;
  isLoading: boolean;
  error: string | null;
}
