import { PostUpstreamRequestBody } from '../types/request_response';

export const CleanupPolicyOptions = [
  {
    short_name: 'LRU 1 month',
    name: 'Least Recently Used (1 Month)',
    code: 'lru_1m',
  },
  {
    short_name: 'LRU 3 months',
    name: 'Least Recently Used (3 Months)',
    code: 'lru_3m',
  },
  {
    short_name: 'Least Pulled',
    name: 'Least Pulled',
    code: 'lp',
  },
];

export const AuthTypeOptions = [
  {
    short_name: 'Anonymous',
    name: 'Anonymous',
    code: 'anonymous',
  },
  {
    short_name: 'Basic Auth',
    name: 'Basic Authentication',
    code: 'basic',
  },
  {
    short_name: 'Bearer',
    name: 'Bearer',
    code: 'bearer',
  },
];

export const UpstreamRegTemplateOptions = [
  {
    name: 'Docker Hub',
    short_name: 'dockerhub',
    code: 'docker-hub',
  },
  {
    name: 'GitHub Container Registry',
    short_name: 'ghcr',
    code: 'github-ghcr',
  },
  {
    name: 'GitLab Container Registry',
    short_name: 'gitlab',
    code: 'gitlab',
  },
  {
    name: 'Amazon Elastic Container Registry',
    short_name: 'ecr',
    code: 'aws-ecr',
  },
  {
    name: 'Google Artifact Registry',
    short_name: 'gcr',
    code: 'google-artifact',
  },
  {
    name: 'Azure Container Registry',
    short_name: 'acr',
    code: 'azure-acr',
  },
  {
    name: 'Harbor',
    short_name: 'harbor',
    code: 'harbor',
  },
  {
    name: 'Quay.io',
    short_name: 'quay',
    code: 'quay-io',
  },
  {
    name: 'JFrog Artifactory',
    short_name: 'jfrog',
    code: 'jfrog-artifactory',
  },
  {
    name: 'Nexus Repository',
    short_name: 'nexus',
    code: 'sonatype-nexus',
  },
  {
    name: 'Other',
    short_name: 'other',
    code: 'other',
  },
];

export const UpstreamTemplates: Record<string, PostUpstreamRequestBody> = {
  'docker-hub': {
    name: 'Docker Hub',
    port: 5000,
    upstream_url: 'https://registry-1.docker.io',
    auth_config: {
      auth_type: 'anonymous',
      credentials_json: null,
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 1024,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 80,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'github-ghcr': {
    name: 'GitHub Container Registry',
    port: 5001,
    upstream_url: 'https://ghcr.io',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        username: '<your-github-username>',
        token: '<your-personal-access-token>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },
  gitlab: {
    name: 'GitLab Container Registry',
    port: 5002,
    upstream_url: 'https://registry.gitlab.com',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        token: '<your-gitlab-token>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'aws-ecr': {
    name: 'Amazon ECR',
    port: 5003,
    upstream_url: '<https://<account>.dkr.ecr.<region>.amazonaws.com>',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        access_key_id: '<AWS_ACCESS_KEY_ID>',
        secret_access_key: '<AWS_SECRET_ACCESS_KEY>',
        region: '<region>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'google-artifact': {
    name: 'Google Artifact Registry',
    port: 5004,
    upstream_url: '<https://<region>-docker.pkg.dev>',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        service_account_json: '<your-service-account-json>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'azure-acr': {
    name: 'Azure Container Registry',
    port: 5005,
    upstream_url: '<https://<your-registry>.azurecr.io>',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        username: '<your-acr-username>',
        password: '<your-acr-password-or-token>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  harbor: {
    name: 'Harbor',
    port: 5006,
    upstream_url: '<https://your-harbor-instance.com>',
    auth_config: {
      auth_type: 'basic',
      credentials_json: {
        username: '<your-username>',
        password: '<your-password>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 1024,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 80,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'quay-io': {
    name: 'Quay.io',
    port: 5007,
    upstream_url: 'https://quay.io',
    auth_config: {
      auth_type: 'bearer',
      credentials_json: {
        token: '<your-quay-token>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 1024,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 80,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'jfrog-artifactory': {
    name: 'JFrog Artifactory',
    port: 5008,
    upstream_url: '<https://<your-domain>.jfrog.io/artifactory/api/docker/<repo-name>/',
    auth_config: {
      auth_type: 'basic',
      credentials_json: {
        username: '<your-username>',
        password: '<your-password>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  'sonatype-nexus': {
    name: 'Nexus Repository',
    port: 5009,
    upstream_url: '<https://your-nexus-host/repository/<repo-name>/',
    auth_config: {
      auth_type: 'basic',
      credentials_json: {
        username: '<your-username>',
        password: '<your-password>',
      },
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 2048,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 85,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },

  other: {
    name: 'Other',
    port: 5010,
    upstream_url: 'https://',
    auth_config: {
      auth_type: 'anonymous',
      credentials_json: null,
    },
    access_config: {
      proxy_enabled: false,
      connection_timeout: 30,
      read_timeout: 60,
      max_connections: 100,
      max_retries: 3,
      retry_delay: 2,
    },
    storage_config: {
      storage_limit: 1024,
      cleanup_policy: 'lru_1m',
      cleanup_threshold: 80,
    },
    cache_config: {
      ttl_seconds: 3600,
      offline_mode: false,
    },
  },
};
