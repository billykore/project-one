import { Post } from "../model";

export interface PostDetailState {
  post: Post | null;
  isLoading: boolean;
  error: string | null;
}
