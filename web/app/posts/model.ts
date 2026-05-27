export interface Post {
  id: number;
  title: string;
  content: string;
  tags?: string[];
  author?: string;
  created_at: string;
  updated_at: string;
  message?: string;
  comments?: Comment[];
}

export interface Comment {
  id: number;
  username: string;
  content: string;
  created_at: string;
}

