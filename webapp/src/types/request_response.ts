import { UserAccountInfo, UserProfileInfo } from "./app_types";

export type UpstreamOCIRegEntity = {
  id?: string;
  name: string;
  url: string;
  port: number;
  status?: string;
  upstream_url: string;
  created_at: string; // ISO string format
  updated_at: string;
};

export type UpstreamOCIRegAuthConfig = {
  auth_type: string;
  credentials_json?: Record<string, any>; // map[string]interface{}
  token_endpoint: string;
  certificate: string;
  updated_at?: string;
};

export type UpstreamOCIRegAccessConfig = {
  proxy_enabled: boolean;
  proxy_url?: string;
  connection_timeout: number;
  read_timeout: number;
  max_connections: number;
  max_retries: number;
  retry_delay: number;
  updated_at: string;
};

export type UpstreamOCIRegStorageConfig = {
  storage_limit: number; // in MBs
  cleanup_policy: string;
  cleanup_threshold: number;
  updated_at: string;
};

export type UpstreamOCIRegCacheConfig = {
  ttl_seconds: number;
  offline_mode: boolean;
  updated_at: string;
};

export type UpstreamOCIRegResMsg = UpstreamOCIRegEntity & {
  auth_config: UpstreamOCIRegAuthConfig;
  access_config: UpstreamOCIRegAccessConfig;
  storage_config: UpstreamOCIRegStorageConfig;
  cache_config: UpstreamOCIRegCacheConfig;
};

export type ListUpstreamRegistriesResponse = {
  total: number;
  page: number;
  limit: number;
  registeries: (UpstreamOCIRegEntity & { cached_images_count: number })[];
};

export type PostUpstreamRequestBody = {
  name: string;
  port: number;
  upstream_url: string;
  auth_config: {
    auth_type: "anonymous" | "basic" | "bearer";
    credentials_json: any;
  };
  access_config: {
    proxy_enabled: boolean;
    connection_timeout: number; // in seconds
    read_timeout: number; // in seconds
    max_connections: number;
    max_retries: number;
    retry_delay: number; // in seconds
  };
  storage_config: {
    storage_limit: number; // in MB
    cleanup_policy: string; // optionally can narrow to `"lru_1m"` | other policies
    cleanup_threshold: number; // in percent
  };
  cache_config: {
    ttl_seconds: number;
    offline_mode: boolean;
  };
};

export type PostUpstreamResponseBody = {
  reg_id?: string | undefined;
  reg_name?: string | undefined;
  error?: string | undefined;
};

export type GetUpstreamResponseBody = UpstreamOCIRegResMsg;

export type SearchImageResponseBody = {
  count: number;
  results: {
    [registery_key: string]: {
      registry_key: string;
      namespace: string;
      image_repository: string;
      tag: string;
      pull_command: string;
    }[];
  };
};

type ImageTagEntity = {
  tag: string;
  pull_command: string;
  last_updated_at: Date;
};

type ImageRepositoryEntity = {
  image_repository: string;
  tags: ImageTagEntity[];
};

type NamespaceEntity = {
  namespace: string;
  image_repositories: ImageRepositoryEntity[];
};

export type ImagesTreeData = {
  default_registry: {
    registry_name: string;
    registry_url: string;
    port: number;
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
    };
    proxy: {
      enable: boolean;
      url: string;
      retries: number;
      socket_timeout: number;
    };
  };
  registeries: {
    registry_name: string;
    registry_url: string;
    port: number;
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
    };
    proxy: {
      enable: boolean;
      url: string;
      retries: number;
      socket_timeout: number;
    };
  }[];
};

export type AuthLoginRequest = {
  username: string;
  password: string;
  scopes: string[];
};

export type AuthLoginResponse = {
  success: boolean;
  error_message: string;
  session_id: string;
  authorized_scopes: string[];
  expires_at: Date;
  user: UserProfileInfo;
};

export type CreateUserAccountRequest = {
  username: string;
  email: string;
  display_name: string;
  role: string;
};

export type CreateUserAccountResponse = {
  username: string;
  user_id: string;
  error?: string;
};

export type UsernameEmailValidationRequest = {
  username: string;
  email: string;
};

export type UsernameEmailValidationResponse = {
  username_available: boolean;
  email_available: boolean;
  error?: string;
};

export type ListUsersResponse = {
  total: number;
  page: number;
  limit: number;
  users: UserAccountInfo[];
  error?: string
}

export type UpdateUserAccountRequest = {
  email: string;
  role: string;
  display_name: string;
}

export type UpdateUserAccountResponse = {
  error?: string;
}

export type UserAccountSetupInfoResponse = {
  error_message: string;
  username: string;
  user_id: string;
  display_name: string;
  email: string;
  role: string;
};

export type PasswordValidationRequest = {
  password: string;
};

export type PasswordValidationResponse = {
  is_valid: boolean;
  msg: string;
};

export type AccountSetupCompleteRequest = {
  user_id: string;
  username: string;
  display_name: string;
  password: string;
  uuid: string;
};

export type CreateNamespaceRequest = {
  name: string;
  description: string;
  is_public: boolean;
  purpose: 'team' | 'project'
  maintainers: string[];
}

export type CreateNamespaceResponse = {
  id: string;
  error_message?: string;
}
