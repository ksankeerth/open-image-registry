import { MenuEntity } from '../types/app_types';
import { flattenMenus } from '../utils/menu';

export const MGMT_CONSOLE_MENUS: MenuEntity[] = [
  {
    name: 'User Management',
    key: 'user-management',
    nav_link: '/console/management',
    description: 'Manage users, invitations, and monitor user activity',
    icon_class: 'pi-users',
    children: [
      {
        name: 'Users',
        key: 'users',
        nav_link: '/console/management/users',
        description: 'Manage user accounts, roles, and permissions',
        children: [],
      },
      {
        name: 'Invitations',
        key: 'invitations',
        nav_link: '/console/management/invitations',
        description: 'Send and manage user invitations to the registry',
        children: [],
      },
      {
        name: 'Activity Log',
        key: 'activity-log',
        nav_link: '/console/management/activity-log',
        description: 'View system activity, user actions, and audit logs',
        children: [],
      },
    ],
  },
  {
    name: 'Resource Management',
    key: 'resource-management',
    nav_link: '/console/management',
    description: 'Configure namespaces, repositories, and upstream registries',
    icon_class: 'pi-shield',
    children: [
      {
        name: 'Namespaces',
        key: 'namespaces',
        nav_link: '/console/management/namespaces',
        description: 'Organize and manage namespaces for your repositories',
        children: [],
      },
      {
        name: 'Repositories',
        key: 'repositories',
        nav_link: '/console/management/repositories',
        description: 'Configure and manage image repositories',
        children: [],
      },
      {
        name: 'Upstreams',
        key: 'upstreams',
        nav_link: '/console/management/upstreams',
        description: 'Configure upstream registry connections for proxy caching',
        children: [],
      },
    ],
  },
  {
    name: 'Integration',
    key: 'integration',
    nav_link: '/console/integration',
    description: 'Configure authentication providers and external integrations',
    icon_class: 'pi-ticket',
    children: [
      {
        name: 'LDAP',
        key: 'ldap',
        nav_link: '/console/integration/ldap',
        description: 'Configure LDAP authentication and user synchronization',
        children: [],
      },
      {
        name: 'SSO',
        key: 'sso',
        nav_link: '/console/integration/sso',
        description: 'Configure Single Sign-On with OAuth2/OIDC providers',
        children: [],
      },
    ],
  },
  {
    name: 'Jobs',
    key: 'jobs',
    nav_link: '/console/jobs',
    description: 'Schedule and monitor automated maintenance tasks',
    icon_class: 'pi-objects-column',
    children: [
      {
        name: 'Cache Cleanup',
        key: 'cache-cleanup',
        nav_link: '/console/jobs/cache-cleanup',
        description: 'Configure and schedule cache cleanup jobs',
        children: [],
      },
      {
        name: 'Storage Cleanup',
        key: 'storage-cleanup',
        nav_link: '/console/jobs/storage-cleanup',
        description: 'Configure and schedule storage cleanup jobs',
        children: [],
      },
      {
        name: 'Manual Jobs',
        key: 'manual-jobs',
        nav_link: '/console/jobs/manual',
        description: 'Run and monitor manual maintenance jobs',
        children: [],
      },
    ],
  },
  {
    name: 'Analytics',
    key: 'analytics',
    nav_link: '/console/analytics',
    description: 'Monitor system performance, usage, and health metrics',
    icon_class: 'pi-wave-pulse',
    children: [
      {
        name: 'Storage',
        key: 'storage-analytics',
        nav_link: '/console/analytics/storage',
        description: 'View storage usage statistics and trends',
        children: [],
      },
      {
        name: 'Cache',
        key: 'cache-analytics',
        nav_link: '/console/analytics/cache',
        description: 'View cache hit rates and performance metrics',
        children: [],
      },
      {
        name: 'Errors',
        key: 'errors-analytics',
        nav_link: '/console/analytics/errors',
        description: 'Monitor and analyze system errors and failures',
        children: [],
      },
      {
        name: 'Garbage Collection',
        key: 'garbage-collection',
        nav_link: '/console/analytics/garbage-collection',
        description: 'View garbage collection status and history',
        children: [],
      },
    ],
  },
];

export const MGMT_MENU_KEY_NAV_LINKS: { key: string; nav_link: string }[] = flattenMenus(
  MGMT_CONSOLE_MENUS
).map((m) => {
  return { key: m.key, nav_link: m.nav_link };
});

export const MGMT_CONSOLE_DEFAULT_PATH: string = '/console/management/users';

export const MGMT_CONSOLE_BASE_PATH: string = '/console/management';

export const REG_CONSOLE_MENUS: MenuEntity[] = [];

export const REG_MENU_KEY_NAV_LINKS: { key: string; nav_link: string }[] = flattenMenus(
  REG_CONSOLE_MENUS
).map((m) => {
  return { key: m.key, nav_link: m.nav_link };
});

export const REG_CONSOLE_DEFAULT_PATH: string = '';

export const REG_CONSOLE_BASE_PATH: string = '/console/registry';
