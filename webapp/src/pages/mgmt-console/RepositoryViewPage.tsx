import { Chip } from 'primereact/chip';
import { Divider } from 'primereact/divider';
import { TabPanel, TabView } from 'primereact/tabview';
import React, { useState } from 'react';
import { RepositoryTagInfo, RepositoryUserAccessInfo, UserAccessInfo } from '../../types/app_types';
import ChipsFilter from '../../components/ChipsFilter';
import { Button } from 'primereact/button';
import { Column } from 'primereact/column';
import { DataTable } from 'primereact/datatable';
import { InputText } from 'primereact/inputtext';
import {
  REPOSITORY_TAG_FILTER_OPTIONS,
  REPOSITORY_USER_ACCESS_FILTER_OPTIONS,
} from '../../config/table_filter';

const mockRepositoryUserAccess: RepositoryUserAccessInfo[] = [
  {
    username: 'alice',
    user_id: 'u001',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'developer',
    granted_by: 'system',
    granted_at: new Date('2025-01-02T09:12:00Z'),
    is_inherited: true,
  },
  {
    username: 'bob',
    user_id: 'u002',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-01',
    access_level: 'developer',
    granted_by: 'alice',
    granted_at: new Date('2025-01-03T10:22:00Z'),
    is_inherited: false,
  },
  {
    username: 'charlie',
    user_id: 'u003',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-02',
    access_level: 'guest',
    granted_by: 'bob',
    granted_at: new Date('2025-01-03T11:10:00Z'),
    is_inherited: false,
  },
  {
    username: 'diana',
    user_id: 'u004',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'developer',
    granted_by: 'alice',
    granted_at: new Date('2025-01-04T14:40:00Z'),
    is_inherited: true,
  },
  {
    username: 'eric',
    user_id: 'u005',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-03',
    access_level: 'guest',
    granted_by: 'alice',
    granted_at: new Date('2025-01-04T09:45:00Z'),
    is_inherited: false,
  },
  {
    username: 'frank',
    user_id: 'u006',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-04',
    access_level: 'developer',
    granted_by: 'bob',
    granted_at: new Date('2025-01-05T08:30:00Z'),
    is_inherited: false,
  },
  {
    username: 'grace',
    user_id: 'u007',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'guest',
    granted_by: 'system',
    granted_at: new Date('2025-01-06T10:18:00Z'),
    is_inherited: true,
  },
  {
    username: 'henry',
    user_id: 'u008',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-05',
    access_level: 'guest',
    granted_by: 'alice',
    granted_at: new Date('2025-01-06T15:20:00Z'),
    is_inherited: false,
  },
  {
    username: 'ivy',
    user_id: 'u009',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-06',
    access_level: 'developer',
    granted_by: 'diana',
    granted_at: new Date('2025-01-07T11:55:00Z'),
    is_inherited: false,
  },
  {
    username: 'jack',
    user_id: 'u010',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'guest',
    granted_by: 'system',
    granted_at: new Date('2025-01-08T09:12:00Z'),
    is_inherited: true,
  },
  {
    username: 'kate',
    user_id: 'u011',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-07',
    access_level: 'developer',
    granted_by: 'alice',
    granted_at: new Date('2025-01-08T13:42:00Z'),
    is_inherited: false,
  },
  {
    username: 'leo',
    user_id: 'u012',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-08',
    access_level: 'guest',
    granted_by: 'diana',
    granted_at: new Date('2025-01-09T07:33:00Z'),
    is_inherited: false,
  },
  {
    username: 'mia',
    user_id: 'u013',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'developer',
    granted_by: 'system',
    granted_at: new Date('2025-01-10T08:00:00Z'),
    is_inherited: true,
  },
  {
    username: 'nick',
    user_id: 'u014',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-09',
    access_level: 'guest',
    granted_by: 'alice',
    granted_at: new Date('2025-01-10T13:10:00Z'),
    is_inherited: false,
  },
  {
    username: 'olivia',
    user_id: 'u015',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-10',
    access_level: 'developer',
    granted_by: 'diana',
    granted_at: new Date('2025-01-11T08:22:00Z'),
    is_inherited: false,
  },
  {
    username: 'paul',
    user_id: 'u016',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'guest',
    granted_by: 'system',
    granted_at: new Date('2025-01-11T10:30:00Z'),
    is_inherited: true,
  },
  {
    username: 'queen',
    user_id: 'u017',
    user_role: 'developer',
    resource_type: 'repository',
    resource_id: 'repo-11',
    access_level: 'developer',
    granted_by: 'alice',
    granted_at: new Date('2025-01-12T09:11:00Z'),
    is_inherited: false,
  },
  {
    username: 'ron',
    user_id: 'u018',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-12',
    access_level: 'guest',
    granted_by: 'bob',
    granted_at: new Date('2025-01-13T14:48:00Z'),
    is_inherited: false,
  },
  {
    username: 'sara',
    user_id: 'u019',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'ns-alpha',
    access_level: 'developer',
    granted_by: 'system',
    granted_at: new Date('2025-01-14T12:00:00Z'),
    is_inherited: true,
  },
  {
    username: 'tom',
    user_id: 'u020',
    user_role: 'guest',
    resource_type: 'repository',
    resource_id: 'repo-13',
    access_level: 'guest',
    granted_by: 'alice',
    granted_at: new Date('2025-01-14T15:55:00Z'),
    is_inherited: false,
  },
];

const mockRepositoryTags: RepositoryTagInfo[] = [
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't001',
    tag: 'v1.0.0',
    last_pushed: new Date('2025-01-10T09:12:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't002',
    tag: 'v1.0.0',
    last_pushed: new Date('2025-01-10T09:12:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't003',
    tag: 'v1.0.1',
    last_pushed: new Date('2025-01-15T11:34:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't004',
    tag: 'v1.0.1',
    last_pushed: new Date('2025-01-15T11:34:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't005',
    tag: 'v1.1.0',
    last_pushed: new Date('2025-01-20T14:28:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't006',
    tag: 'v1.1.0',
    last_pushed: new Date('2025-01-20T14:28:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't007',
    tag: 'v1.1.0-rc1',
    last_pushed: new Date('2025-01-18T10:10:00Z'),
    stable: false,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't008',
    tag: 'v1.1.0-rc1',
    last_pushed: new Date('2025-01-18T10:10:00Z'),
    stable: false,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't009',
    tag: 'v1.2.0',
    last_pushed: new Date('2025-02-01T07:55:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't010',
    tag: 'v1.2.0',
    last_pushed: new Date('2025-02-01T07:55:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't011',
    tag: 'v1.2.1',
    last_pushed: new Date('2025-02-03T10:05:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't012',
    tag: 'v1.2.1',
    last_pushed: new Date('2025-02-03T10:05:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't013',
    tag: 'v1.3.0',
    last_pushed: new Date('2025-02-08T12:44:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't014',
    tag: 'v1.3.0',
    last_pushed: new Date('2025-02-08T12:44:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't015',
    tag: 'latest',
    last_pushed: new Date('2025-02-08T12:44:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't016',
    tag: 'latest',
    last_pushed: new Date('2025-02-08T12:44:00Z'),
    stable: true,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't017',
    tag: 'debug',
    last_pushed: new Date('2025-02-09T09:30:00Z'),
    stable: false,
    platform_os: 'linux',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't018',
    tag: 'debug',
    last_pushed: new Date('2025-02-09T09:30:00Z'),
    stable: false,
    platform_os: 'linux',
    platform_arch: 'arm64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't019',
    tag: 'win-v1.0.0',
    last_pushed: new Date('2025-01-20T14:55:00Z'),
    stable: false,
    platform_os: 'windows',
    platform_arch: 'amd64',
  },
  {
    namespace_id: 'team-platform-eng',
    repository_id: 'order-backend',
    tag_id: 't020',
    tag: 'win-v1.0.0',
    last_pushed: new Date('2025-01-20T14:55:00Z'),
    stable: false,
    platform_os: 'windows',
    platform_arch: 'arm64',
  },
];

const RepositoryViewPage = () => {
  const [userAccesList, setUserAccessList] =
    useState<RepositoryUserAccessInfo[]>(mockRepositoryUserAccess);
  const [tagsList, setTagsList] = useState<RepositoryTagInfo[]>(mockRepositoryTags);

  const handleUserAccessFilter = (options: string[]): void => { };

  const handleTagFilter = (options: string[]): void => { };

  return (
    <div className="flex flex-column p-2 pt-4 gap-3 ">
      <div className="bg-white border-round-lg flex flex-column gap-2">
        <div className="flex flex-column pt-2 gap-2">
          <div className="flex flex-row justify-content-between pt-2 pb-2">
            <div className="pl-3">
              <span className="font-semibold text-lg">Repository:</span>
              &nbsp;&nbsp;&nbsp;
              <span className=" font-medium text-lg">
                <span className="cursor-pointer underline text-blue-600">team-platform-eng</span>{' '}
                <span className="text-sm">/</span>order-backend
              </span>
              &nbsp;&nbsp;
              <Chip
                className="pl-2 pr-2 p-0  bg-white border-600 border-1 border-teal-300"
                style={{ fontSize: '0.60rem' }}
                label="active"
              />
              &nbsp;&nbsp;
              <Chip
                className="pl-2 pr-2 p-0  bg-white border-400 border-1"
                style={{ fontSize: '0.60rem' }}
                label="public"
              />
            </div>

            <div className="flex gap-2 pr-3">
              <span className="pi pi-ellipsis-v"></span>
            </div>
          </div>
          <div className="flex flex-row justify-content-between align-items-center gap-4 pl-3 w-full">
            <div className="flex-grow-0 flex flex-row gap-4 mr-3">
              <div>
                <span className="pi pi-clock text-xs"></span>
                &nbsp;&nbsp;
                <span className="text-xs">{new Date().toUTCString()}</span>
              </div>
              <div>
                <span className="pi pi-user text-xs"></span>
                &nbsp;&nbsp;
                <span className="text-xs">admin</span>
              </div>
              <div>
                <span className="pi pi-tag text-xs"></span>
                &nbsp;&nbsp;
                <span className="text-xs">64</span>
              </div>
            </div>
          </div>
          <Divider layout="horizontal" className="mt-0 pt-0" />
          <div className="flex flex-row pl-3 pr-3 pb-2">
            <span className="font-semibold">
              Access Inheritance & Rules&nbsp;&nbsp;
              <span className="cursor-pointer pi pi-info-circle text-xs"></span>
            </span>
          </div>
          <div className="flex flex-column w-full gap-3 pl-4 pr-4">
            <div className="flex w-full justify-content-between flex-row flex-grow-1 gap-5">
              <div className="w-5 flex justify-content-center text-sm font-medium">
                <span className="font-semibold text-sm">Namespace Level</span>
              </div>

              <div className="w-9 flex justify-content-center  text-sm font-medium">
                <span className="font-semibold text-sm">Repository Level</span>
              </div>
            </div>

            <div className="flex w-full justify-content-between align-items-center  flex-row flex-grow-1 gap-5">
              <div className="flex flex-column gap-2 max-w-28rem w-5  border-primary-500 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-primary-700">
                  <span className="pi pi-shield font-semibold" />
                  &nbsp;&nbsp;&nbsp;
                  <span className="font-semibold">Maintainer</span>
                </div>
                <div className="text-sm">Full control over namespace and all repositories.</div>
              </div>
              <div className="flex justify-content-center align-items-center text-primary-400">
                <span className="pi pi-chevron-right"></span>
              </div>
              <div className="flex flex-column gap-2 max-w-28rem w-9  border-red-500 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-red-400">
                  <span className="pi pi-times font-semibold text-xs" />
                  &nbsp;&nbsp;
                  <span className="font-semibold text-sm">Cannot Override</span>
                </div>
                <div className="text-sm">
                  Automatically has full access to all repositories. No repository-level assignment
                  needed or allowed.{' '}
                </div>
              </div>
            </div>

            <div className="flex w-full justify-content-between align-items-center flex-row flex-grow-1 gap-5">
              <div className="flex flex-column gap-2 max-w-28rem w-5  border-primary-400 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-primary-600">
                  <span className="pi pi-lock font-semibold" />
                  &nbsp;&nbsp;&nbsp;
                  <span className="font-semibold">Developer</span>
                </div>
                <div className="text-sm">Can read, write, and manage repositories.</div>
              </div>
              <div className="flex justify-content-center align-items-center text-primary-400">
                <span className="pi pi-chevron-right"></span>
              </div>
              <div className="flex flex-column gap-2 max-w-28rem w-9  border-red-500 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-red-400">
                  <span className="pi pi-times font-semibold text-xs" />
                  &nbsp;&nbsp;
                  <span className="font-semibold text-sm">Cannot Override</span>
                </div>
                <div className="text-sm">
                  Automatically has developer access to all repositories. Cannot be changed at
                  repository level.
                </div>
              </div>
            </div>

            <div className="flex w-full justify-content-between align-items-center flex-row flex-grow-1 gap-5">
              <div className="flex flex-column  gap-2 max-w-28rem w-5 border-orange-400 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-orange-600">
                  <span className="pi pi-eye font-semibold" />
                  &nbsp;&nbsp;&nbsp;
                  <span className="font-semibold">Guest</span>
                </div>
                <div className="text-sm flex flex-column">
                  <div>Read-only access to repositories.</div>
                </div>
              </div>
              <div className="flex justify-content-center align-items-center text-orange-300">
                <span className="pi pi-chevron-right"></span>
              </div>
              <div className="flex flex-column gap-2 max-w-28rem w-9  border-primary-400 border-1 border-round-xl p-2 pl-3">
                <div className="font-semibold text-primary-600">
                  <span className="pi pi-check font-semibold text-xs" />
                  &nbsp;&nbsp;
                  <span className="font-semibold text-sm">Can Upgrade</span>
                </div>
                <div className="text-sm flex flex-column">
                  <div>
                    Can be granted <span className="font-semibold">Developer</span> access to
                    specific repositories.
                  </div>
                </div>
              </div>
            </div>

            <div className="flex  surface-50 border-left-3 gap-2 m-1 p-2 border-round-lg">
              <div className="flex align-items-start">
                <div className="flex align-items-center">
                  <span className="pi pi-lightbulb  text-yellow-400 bg-yellow-100"></span>
                  &nbsp;
                  <span className="text-sm font-semibold">Note:</span>
                </div>
              </div>

              <span className="text-sm flex-grow-1">
                Repository level only supports Developer and Guest access levels. Maintainer access
                is not available at the repository level. Additionally, user access level at
                namespace cannot be downgraded at repository level (e.g., Developer at namespace
                cannot become Guest at repository).
              </span>
            </div>
          </div>
          <TabView>
            <TabPanel header={<div className="font-medium">User Access</div>}>
              <div className="flex flex-row align-items-center justify-content-between pl-3 pr-2 ">
                <div className="flex align-items-center  flex-grow-1">
                  <div className=" text-sm"> Select Filter</div>
                  <ChipsFilter
                    filterOptions={REPOSITORY_USER_ACCESS_FILTER_OPTIONS}
                    handleFilterChange={handleUserAccessFilter}
                  />
                </div>
                <Button
                  size="small"
                  outlined
                  className="p-0 m-0 border-1 border-solid border-round-lg border-teal-100  text-xs flex-grow-0 w-3"
                >
                  <InputText
                    size="10"
                    type="text"
                    className="border-none p-2 text-sm "
                    placeholder="Search users . . . ."
                  />
                  <i className="pi pi-search text-teal-400 text-sm pr-2 "></i>
                </Button>
              </div>
              <DataTable
                value={userAccesList}
                paginator
                paginatorLeft={<div className="text-xs">82 users found</div>}
                rows={8}
                totalRecords={mockRepositoryUserAccess.length}
              >
                <Column field="username" header="User" sortable />
                <Column
                  field="access_level"
                  header="Access Level"
                  body={(data: RepositoryUserAccessInfo) => (
                    <div>
                      <span>{data.access_level}</span>
                      {data.is_inherited && (
                        <span>
                          {' '}
                          &nbsp;
                          <Chip
                            className="pl-2 pr-2 p-0  bg-white border-400 border-1"
                            style={{ fontSize: '0.60rem' }}
                            label="inherited"
                          />
                        </span>
                      )}
                    </div>
                  )}
                />

                <Column header="Granted By" field="granted_by" sortable />
                <Column
                  header="Granted At"
                  field="granted_at"
                  body={(data: UserAccessInfo) => data.granted_at.toUTCString()}
                  sortable
                />
                <Column
                  align="right"
                  header={
                    <div>
                      <Button
                        size="small"
                        severity="warning"
                        className="p-2 m-0  border-1 border-solid border-round-lg border-teal-100 text-xs"
                      >
                        <span>Grant Access</span>
                      </Button>
                    </div>
                  }
                  body={(data) => {
                    return (
                      <div className="flex justify-content-end">
                        <span className="pi pi-ellipsis-v text-sm"></span>
                      </div>
                    );
                  }}
                />
              </DataTable>
            </TabPanel>
            <TabPanel header={<div className="font-medium">Tags</div>}>
              <div className="flex flex-row align-items-center justify-content-between pl-3 pr-2 ">
                <div className="flex align-items-center  flex-grow-1">
                  <div className=" text-sm"> Select Filter</div>
                  <ChipsFilter
                    filterOptions={REPOSITORY_TAG_FILTER_OPTIONS}
                    handleFilterChange={handleTagFilter}
                    maxChipsPerRow={5}
                  />
                </div>
                <Button
                  size="small"
                  outlined
                  className="p-0 m-0 border-1 border-solid border-round-lg border-teal-100  text-xs flex-grow-0 w-3"
                >
                  <InputText
                    size="10"
                    type="text"
                    className="border-none p-2 text-sm "
                    placeholder="Search tags, OS, architecture . . . ."
                  />
                  <i className="pi pi-search text-teal-400 text-sm pr-2 "></i>
                </Button>
              </div>
              <DataTable
                value={tagsList}
                paginator
                paginatorLeft={<div className="text-xs">82 tags found</div>}
                rows={8}
                totalRecords={mockRepositoryTags.length}
              >
                <Column
                  field="tag"
                  header="Tag"
                  body={(data: RepositoryTagInfo) => (
                    <div>
                      <span>{data.tag}</span> &nbsp;
                      {data.stable && (
                        <Chip
                          className="pl-2 pr-2 p-0  bg-white border-400 border-1"
                          style={{ fontSize: '0.60rem' }}
                          label="stable"
                        />
                      )}
                    </div>
                  )}
                  sortable
                />
                <Column
                  header="Platform"
                  body={(data: RepositoryTagInfo) => (
                    <div>
                      <span>{data.platform_arch}</span> &nbsp;/ <span>{data.platform_os}</span>
                    </div>
                  )}
                />

                <Column
                  header="Pushed At"
                  field="last_pushed"
                  body={(data: RepositoryTagInfo) => data.last_pushed.toUTCString()}
                  sortable
                />
                <Column
                  align="right"
                  body={(data) => {
                    return (
                      <div className="flex justify-content-end">
                        <span className="pi pi-ellipsis-v text-sm"></span>
                      </div>
                    );
                  }}
                />
              </DataTable>
            </TabPanel>
          </TabView>
        </div>
      </div>
    </div>
  );
};

export default RepositoryViewPage;
