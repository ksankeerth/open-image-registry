export enum RegistryTypes {
  Hosted,
  Proxy,
}

export enum ImageTagsTreeNodeTypes {
  Registry,
  Namespace,
  Repository,
  Tag,
}

export type ImageTagsTreeViewNodeData = {
  registry_key: string;
  registry_type: RegistryTypes;
  image_or_tag_path: string;
  node_type: ImageTagsTreeNodeTypes;
  namespace: string;
  image_repository: string;
  tag: string;
};

export type NamespaceAccess = {
  id: string;
  namespace: string;
  resource_id: string;
  user_id: string;
  access_level: string;
  granted_by: string;
  created_at: Date;
  updated_at: Date;
};

export type RepositoryAccesss = {
  id: string;
  namespace: string;
  repository: string;
  resource_id: string;
  user_id: string;
  access_level: string;
  granted_by: string;
  created_at: Date;
  updated_at: Date;
};

export type UserProfileInfo = {
  user_id: string;
  username: string;
  role: string;
  namespaces: NamespaceAccess[];
  repositories: RepositoryAccesss[];
};

export type UserAccountInfo = {
  id: string;
  username: string;
  role: string;
  email: string;
  display_name: string;
  locked: boolean;
  locked_reason: string;
  failed_attempts: number;
  locked_at?: Date;
  password_recovery_id: string;
  password_recovery_reason: string;
  password_recovery_at?: Date;
  last_loggedin_at?: Date;
  created_at: Date;
  updated_at?: Date;
};

export type TableSortState<T> = {
  key: keyof T;
  order: -1 | 0 | 1;
};

export type TablePaginationState = {
  page: number;
  limit: number;
};

export type TableColumnFilterState<T, K extends keyof T = keyof T> = {
  key: keyof T;
  values: T[K][]; // equal filter only for simplicity
}

export type TableFilterSearchPaginationSortState<T> = {
  sort?: TableSortState<T>;
  pagination: TablePaginationState;
  filters?: TableColumnFilterState<T>[];
  search_value?: string;
};


export type MenuItem = {
  name: string;
  key: string;
  description: string;
  nav_link: string;
  icon_class?: string;
  collapsed?: boolean;
  children: MenuItem[];
};

type UserAccess = {
  maintainers: number;
  guests: number;
  developers: number;
};

type RepositoryStats = {
  total: number;
  restricted: number;
  deprecated: number;
  active: number;
  disabled: number;
};

export type NamespaceInfo = {
  id: string;
  name: string;
  is_public: boolean;
  type: 'project' | 'team';
  state: 'active' | 'deprecated' | 'disabled';
  description: string;
  created_at: Date;
  created_by: string;
  updated_at?: Date;
  user_access: UserAccess;
  repositories: RepositoryStats;
};

export type UserAccessInfo = {
  username: string;
  user_id: string;
  user_role: 'maintainer' | 'developer' | 'guest'
  resource_type: 'namespace' | 'repository' | 'upstream_registry'
  resource_id: string;
  access_level: 'maintainer' | 'developer' | 'guest'
  granted_by: string;
  granted_at: Date;
}

export type RepositoryUserAccessInfo = UserAccessInfo & { is_inherited: boolean };

export type NamespaceRepositoryInfo = {
  name: string;
  id: string;
  namesapce: string;
  namespace_id: string;
  created_at: Date;
  created_by: string;
  tags_count: string;
  state: 'active' | 'deprecated' | 'disabled';
}

export type ChangeTrackerEventInfo = {
  id: string;
  timestamp: Date;
  type: 'add' | 'change' | 'delete';
  message: string;
}

export type RepositoryTagInfo = {
  namespace_id: string;
  repository_id: string;
  tag_id: string;
  tag: string;
  last_pushed: Date;
  stable: boolean;
  platform_os: string;
  platform_arch: string;
}