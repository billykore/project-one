export interface SearchResult {
  username: string;
  name: string;
}

export interface SearchResponse {
  data: SearchResult[];
  next_cursor: string;
  has_more: boolean;
}
