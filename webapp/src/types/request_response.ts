export type PostUpstreamRequestBody = {
  registry_name : string;
  registry_url: string;
  auth: {
    username: string;
    password: string;
  };
  storage: {
    limit: number;
    cleanup_threshold: number;
  };
  cache: {
    ttl: number;
    offline_mode: boolean;
  },
  proxy: {
    enable: boolean; 
    url: string;
    retries: number;
    socket_timeout: number;
  }
}