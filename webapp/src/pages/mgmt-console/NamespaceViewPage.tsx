import { Button } from 'primereact/button';
import { Chip } from 'primereact/chip';
import { TabPanel, TabView } from 'primereact/tabview';
import React, { useState } from 'react';
import {
  ChangeTrackerEventInfo,
  NamespaceRepositoryInfo,
  UserAccessInfo,
} from '../../types/app_types';
import ChangeTrackerView from '../../components/ChangeTrackerView';
import { Divider } from 'primereact/divider';
import { DataView } from 'primereact/dataview';
import ChipsFilter from '../../components/ChipsFilter';
import { REPOSITORY_FILTER_OPTIONS, USER_ACCESS_FILTER_OPTIONS } from '../../config/table_filter';
import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { SplitButton } from 'primereact/splitbutton';
import { InputText } from 'primereact/inputtext';
import { useNavigate } from 'react-router-dom';

const namespaceMaintainers: UserAccessInfo[] = [
  {
    username: 'john',
    user_id: 'u001',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'openimage-core',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2024-01-01T10:00:00Z'),
  },
  {
    username: 'sara',
    user_id: 'u002',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'platform-services',
    access_level: 'maintainer',
    granted_by: 'john',
    granted_at: new Date('2024-01-03T11:44:00Z'),
  },
  {
    username: 'kevin',
    user_id: 'u003',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'frontend-ui',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2023-12-22T09:30:00Z'),
  },
  {
    username: 'emma',
    user_id: 'u004',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'devtools',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2024-03-15T13:12:00Z'),
  },
  {
    username: 'alex',
    user_id: 'u005',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'backend-services',
    access_level: 'maintainer',
    granted_by: 'kate',
    granted_at: new Date('2024-02-12T08:24:00Z'),
  },
  {
    username: 'olivia',
    user_id: 'u006',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'ai-ml',
    access_level: 'maintainer',
    granted_by: 'john',
    granted_at: new Date('2024-01-20T07:40:00Z'),
  },
  {
    username: 'dylan',
    user_id: 'u007',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'analytics',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2023-11-10T11:00:00Z'),
  },
  {
    username: 'sophia',
    user_id: 'u008',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'qa-testing',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2023-10-09T15:10:00Z'),
  },
  {
    username: 'liam',
    user_id: 'u009',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'proxy-cache',
    access_level: 'maintainer',
    granted_by: 'john',
    granted_at: new Date('2023-09-21T15:45:00Z'),
  },
  {
    username: 'noah',
    user_id: 'u010',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'infra-ops',
    access_level: 'maintainer',
    granted_by: 'sara',
    granted_at: new Date('2023-08-14T09:51:00Z'),
  },
  {
    username: 'ella',
    user_id: 'u011',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'openimage-examples',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2023-09-02T10:01:00Z'),
  },
  {
    username: 'tony',
    user_id: 'u012',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'frontend-utils',
    access_level: 'maintainer',
    granted_by: 'kevin',
    granted_at: new Date('2023-07-18T13:44:00Z'),
  },
  {
    username: 'mike',
    user_id: 'u013',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'render-engine',
    access_level: 'maintainer',
    granted_by: 'emma',
    granted_at: new Date('2024-01-10T10:10:00Z'),
  },
  {
    username: 'ray',
    user_id: 'u014',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'image-optimizer',
    access_level: 'maintainer',
    granted_by: 'emma',
    granted_at: new Date('2024-03-10T10:10:00Z'),
  },
  {
    username: 'adam',
    user_id: 'u015',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'workflows',
    access_level: 'maintainer',
    granted_by: 'kate',
    granted_at: new Date('2024-01-22T10:10:00Z'),
  },
  {
    username: 'amy',
    user_id: 'u016',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'deploy-ops',
    access_level: 'maintainer',
    granted_by: 'john',
    granted_at: new Date('2023-12-30T10:10:00Z'),
  },
  {
    username: 'rick',
    user_id: 'u017',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'cluster-pipeline',
    access_level: 'maintainer',
    granted_by: 'kevin',
    granted_at: new Date('2024-02-20T11:10:00Z'),
  },
  {
    username: 'karen',
    user_id: 'u018',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'ai-vision',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2023-11-14T09:10:00Z'),
  },
  {
    username: 'lucas',
    user_id: 'u019',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'preprocessing',
    access_level: 'maintainer',
    granted_by: 'admin',
    granted_at: new Date('2024-02-14T09:10:00Z'),
  },
  {
    username: 'bella',
    user_id: 'u020',
    user_role: 'maintainer',
    resource_type: 'namespace',
    resource_id: 'model-objects',
    access_level: 'maintainer',
    granted_by: 'john',
    granted_at: new Date('2024-03-10T09:10:00Z'),
  },
];

const namespaceDevelopers: UserAccessInfo[] = [
  {
    username: 'tim',
    user_id: 'u021',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'frontend-ui',
    access_level: 'developer',
    granted_by: 'kevin',
    granted_at: new Date('2023-12-02T08:51:00Z'),
  },
  {
    username: 'maya',
    user_id: 'u022',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'qa-testing',
    access_level: 'developer',
    granted_by: 'emma',
    granted_at: new Date('2024-01-06T08:10:00Z'),
  },
  {
    username: 'robert',
    user_id: 'u023',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'backend-services',
    access_level: 'developer',
    granted_by: 'alex',
    granted_at: new Date('2024-03-06T11:10:00Z'),
  },
  {
    username: 'dan',
    user_id: 'u024',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'ai-vision',
    access_level: 'developer',
    granted_by: 'karen',
    granted_at: new Date('2023-12-28T07:10:00Z'),
  },
  {
    username: 'jane',
    user_id: 'u025',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'openimage-core',
    access_level: 'developer',
    granted_by: 'john',
    granted_at: new Date('2024-01-15T07:10:00Z'),
  },
  {
    username: 'mark',
    user_id: 'u026',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'devtools',
    access_level: 'developer',
    granted_by: 'alex',
    granted_at: new Date('2024-02-10T09:10:00Z'),
  },
  {
    username: 'eldon',
    user_id: 'u027',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'render-engine',
    access_level: 'developer',
    granted_by: 'mike',
    granted_at: new Date('2023-12-10T10:10:00Z'),
  },
  {
    username: 'harry',
    user_id: 'u028',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'infra-ops',
    access_level: 'developer',
    granted_by: 'noah',
    granted_at: new Date('2023-09-10T10:10:00Z'),
  },
  {
    username: 'steve',
    user_id: 'u029',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'proxy-cache',
    access_level: 'developer',
    granted_by: 'liam',
    granted_at: new Date('2023-10-10T11:10:00Z'),
  },
  {
    username: 'anna',
    user_id: 'u030',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'cluster-pipeline',
    access_level: 'developer',
    granted_by: 'rick',
    granted_at: new Date('2024-03-10T11:10:00Z'),
  },
  {
    username: 'cam',
    user_id: 'u031',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'platform-services',
    access_level: 'developer',
    granted_by: 'sara',
    granted_at: new Date('2024-02-10T10:10:00Z'),
  },
  {
    username: 'irene',
    user_id: 'u032',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'analytics',
    access_level: 'developer',
    granted_by: 'dylan',
    granted_at: new Date('2024-02-14T10:10:00Z'),
  },
  {
    username: 'oscar',
    user_id: 'u033',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'openimage-examples',
    access_level: 'developer',
    granted_by: 'ella',
    granted_at: new Date('2023-11-11T10:10:00Z'),
  },
  {
    username: 'natalie',
    user_id: 'u034',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'model-objects',
    access_level: 'developer',
    granted_by: 'bella',
    granted_at: new Date('2023-12-10T10:10:00Z'),
  },
  {
    username: 'raymond',
    user_id: 'u035',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'image-optimizer',
    access_level: 'developer',
    granted_by: 'ray',
    granted_at: new Date('2024-03-01T10:10:00Z'),
  },
  {
    username: 'jason',
    user_id: 'u036',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'frontend-utils',
    access_level: 'developer',
    granted_by: 'tony',
    granted_at: new Date('2023-10-04T11:10:00Z'),
  },
  {
    username: 'becky',
    user_id: 'u037',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'deploy-ops',
    access_level: 'developer',
    granted_by: 'amy',
    granted_at: new Date('2023-12-20T09:10:00Z'),
  },
  {
    username: 'ivan',
    user_id: 'u038',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'workflows',
    access_level: 'developer',
    granted_by: 'adam',
    granted_at: new Date('2024-03-05T09:10:00Z'),
  },
  {
    username: 'timothy',
    user_id: 'u039',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'preprocessing',
    access_level: 'developer',
    granted_by: 'lucas',
    granted_at: new Date('2024-03-07T09:10:00Z'),
  },
  {
    username: 'lara',
    user_id: 'u040',
    user_role: 'developer',
    resource_type: 'namespace',
    resource_id: 'qa-testing',
    access_level: 'developer',
    granted_by: 'sophia',
    granted_at: new Date('2024-03-09T10:10:00Z'),
  },
];

const namespaceGuests: UserAccessInfo[] = [
  {
    username: 'userA',
    user_id: 'u041',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'openimage-core',
    access_level: 'guest',
    granted_by: 'john',
    granted_at: new Date('2023-11-05T09:44:00Z'),
  },
  {
    username: 'userB',
    user_id: 'u042',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'qa-testing',
    access_level: 'guest',
    granted_by: 'emma',
    granted_at: new Date('2024-01-02T09:44:00Z'),
  },
  {
    username: 'userC',
    user_id: 'u043',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'ai-ml',
    access_level: 'guest',
    granted_by: 'olivia',
    granted_at: new Date('2023-12-03T09:44:00Z'),
  },
  {
    username: 'userD',
    user_id: 'u044',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'proxy-cache',
    access_level: 'guest',
    granted_by: 'liam',
    granted_at: new Date('2023-11-04T11:44:00Z'),
  },
  {
    username: 'userE',
    user_id: 'u045',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'backend-services',
    access_level: 'guest',
    granted_by: 'alex',
    granted_at: new Date('2023-11-14T11:44:00Z'),
  },
  {
    username: 'userF',
    user_id: 'u046',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'frontend-ui',
    access_level: 'guest',
    granted_by: 'kevin',
    granted_at: new Date('2023-11-24T09:44:00Z'),
  },
  {
    username: 'userG',
    user_id: 'u047',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'devtools',
    access_level: 'guest',
    granted_by: 'emma',
    granted_at: new Date('2024-01-08T09:44:00Z'),
  },
  {
    username: 'userH',
    user_id: 'u048',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'analytics',
    access_level: 'guest',
    granted_by: 'dylan',
    granted_at: new Date('2023-10-18T09:44:00Z'),
  },
  {
    username: 'userI',
    user_id: 'u049',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'render-engine',
    access_level: 'guest',
    granted_by: 'mike',
    granted_at: new Date('2024-02-10T09:44:00Z'),
  },
  {
    username: 'userJ',
    user_id: 'u050',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'workflows',
    access_level: 'guest',
    granted_by: 'adam',
    granted_at: new Date('2023-10-14T11:44:00Z'),
  },
  {
    username: 'userK',
    user_id: 'u051',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'preprocessing',
    access_level: 'guest',
    granted_by: 'lucas',
    granted_at: new Date('2024-02-11T11:44:00Z'),
  },
  {
    username: 'userL',
    user_id: 'u052',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'model-objects',
    access_level: 'guest',
    granted_by: 'bella',
    granted_at: new Date('2023-11-21T10:44:00Z'),
  },
  {
    username: 'userM',
    user_id: 'u053',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'image-optimizer',
    access_level: 'guest',
    granted_by: 'ray',
    granted_at: new Date('2023-12-09T10:44:00Z'),
  },
  {
    username: 'userN',
    user_id: 'u054',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'team-collab',
    access_level: 'guest',
    granted_by: 'john',
    granted_at: new Date('2024-01-14T10:44:00Z'),
  },
  {
    username: 'userO',
    user_id: 'u055',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'cluster-pipeline',
    access_level: 'guest',
    granted_by: 'rick',
    granted_at: new Date('2023-11-10T08:44:00Z'),
  },
  {
    username: 'userP',
    user_id: 'u056',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'deploy-ops',
    access_level: 'guest',
    granted_by: 'amy',
    granted_at: new Date('2024-02-10T08:44:00Z'),
  },
  {
    username: 'userQ',
    user_id: 'u057',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'platform-services',
    access_level: 'guest',
    granted_by: 'sara',
    granted_at: new Date('2023-12-20T08:44:00Z'),
  },
  {
    username: 'userR',
    user_id: 'u058',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'infra-ops',
    access_level: 'guest',
    granted_by: 'noah',
    granted_at: new Date('2023-11-05T08:44:00Z'),
  },
  {
    username: 'userS',
    user_id: 'u059',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'frontend-utils',
    access_level: 'guest',
    granted_by: 'tony',
    granted_at: new Date('2023-11-09T08:44:00Z'),
  },
  {
    username: 'userT',
    user_id: 'u060',
    user_role: 'guest',
    resource_type: 'namespace',
    resource_id: 'openimage-examples',
    access_level: 'guest',
    granted_by: 'ella',
    granted_at: new Date('2023-11-11T08:44:00Z'),
  },
];

export const mockEvents: ChangeTrackerEventInfo[] = [
  {
    id: '1',
    timestamp: new Date('2025-01-02T09:12:00Z'),
    type: 'add',
    message: "Namespace 'ns-alpha' created",
  },
  {
    id: '2',
    timestamp: new Date('2025-01-02T10:42:00Z'),
    type: 'add',
    message: "Repository 'ns-alpha/core' created",
  },
  {
    id: '3',
    timestamp: new Date('2025-01-03T07:20:00Z'),
    type: 'add',
    message: "Developer added to namespace 'ns-alpha'",
  },
  {
    id: '4',
    timestamp: new Date('2025-01-05T13:10:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' visibility changed from public to private",
  },
  {
    id: '5',
    timestamp: new Date('2025-01-06T09:55:00Z'),
    type: 'add',
    message: "Maintainer assigned to namespace 'ns-alpha'",
  },
  {
    id: '6',
    timestamp: new Date('2025-01-06T10:07:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-alpha'",
  },
  {
    id: '7',
    timestamp: new Date('2025-01-07T11:16:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' lifecycle changed from active to deprecated",
  },
  {
    id: '8',
    timestamp: new Date('2025-01-08T14:33:00Z'),
    type: 'add',
    message: "Repository 'ns-alpha/cli' created",
  },
  {
    id: '9',
    timestamp: new Date('2025-01-08T14:34:00Z'),
    type: 'change',
    message: "Repository 'ns-alpha/cli' restricted to selected users",
  },
  {
    id: '10',
    timestamp: new Date('2025-01-08T15:40:00Z'),
    type: 'add',
    message: "Tag 'v1.0.0' of 'ns-alpha/core' marked as stable",
  },

  {
    id: '11',
    timestamp: new Date('2025-01-12T11:25:00Z'),
    type: 'change',
    message: "Tag 'v1.0.0' of 'ns-alpha/core' marked as unstable",
  },
  {
    id: '12',
    timestamp: new Date('2025-01-14T07:20:00Z'),
    type: 'delete',
    message: "Repository 'ns-alpha/cli' was deleted",
  },
  {
    id: '13',
    timestamp: new Date('2025-01-15T09:12:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' lifecycle changed from deprecated to disabled",
  },
  {
    id: '14',
    timestamp: new Date('2025-01-17T08:42:00Z'),
    type: 'delete',
    message: "Maintainer removed from namespace 'ns-alpha'",
  },
  {
    id: '15',
    timestamp: new Date('2025-01-18T10:01:00Z'),
    type: 'add',
    message: "Namespace 'ns-prod' created",
  },
  {
    id: '16',
    timestamp: new Date('2025-01-18T10:33:00Z'),
    type: 'add',
    message: "Repository 'ns-prod/app' created",
  },
  {
    id: '17',
    timestamp: new Date('2025-01-19T09:20:00Z'),
    type: 'add',
    message: "Developer added to namespace 'ns-prod'",
  },
  {
    id: '18',
    timestamp: new Date('2025-01-20T11:12:00Z'),
    type: 'add',
    message: "Maintainer assigned to namespace 'ns-prod'",
  },
  {
    id: '19',
    timestamp: new Date('2025-01-21T09:55:00Z'),
    type: 'change',
    message: "Repository 'ns-prod/app' lifecycle changed to deprecated",
  },
  {
    id: '20',
    timestamp: new Date('2025-01-22T14:33:00Z'),
    type: 'add',
    message: "Tag 'v0.9.0' of 'ns-prod/app' marked as stable",
  },

  {
    id: '21',
    timestamp: new Date('2025-01-23T09:02:00Z'),
    type: 'change',
    message: "Repository 'ns-prod/app' lifecycle changed to disabled",
  },
  {
    id: '22',
    timestamp: new Date('2025-01-24T07:10:00Z'),
    type: 'delete',
    message: "Repository 'ns-prod/app' was deleted",
  },
  {
    id: '23',
    timestamp: new Date('2025-01-25T12:45:00Z'),
    type: 'add',
    message: "Namespace 'ns-qa' created",
  },
  {
    id: '24',
    timestamp: new Date('2025-01-25T13:15:00Z'),
    type: 'add',
    message: "Repository 'ns-qa/tests' created",
  },
  {
    id: '25',
    timestamp: new Date('2025-01-26T10:27:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-qa'",
  },
  {
    id: '26',
    timestamp: new Date('2025-01-27T15:20:00Z'),
    type: 'change',
    message: "Namespace 'ns-qa' visibility changed from private to public",
  },
  {
    id: '27',
    timestamp: new Date('2025-01-28T09:10:00Z'),
    type: 'change',
    message: "Tag 'v2.0.0' of 'ns-alpha/core' marked as stable",
  },
  {
    id: '28',
    timestamp: new Date('2025-01-29T11:11:00Z'),
    type: 'change',
    message: "Tag 'v2.0.0' of 'ns-alpha/core' marked as unstable",
  },
  {
    id: '29',
    timestamp: new Date('2025-01-30T08:42:00Z'),
    type: 'add',
    message: "Repository 'ns-alpha/ui' created",
  },
  {
    id: '30',
    timestamp: new Date('2025-02-01T09:12:00Z'),
    type: 'add',
    message: "Developer added to namespace 'ns-alpha'",
  },

  {
    id: '31',
    timestamp: new Date('2025-02-02T10:55:00Z'),
    type: 'add',
    message: "Namespace 'ns-dev' created",
  },
  {
    id: '32',
    timestamp: new Date('2025-02-02T12:20:00Z'),
    type: 'add',
    message: "Maintainer assigned to namespace 'ns-dev'",
  },
  {
    id: '33',
    timestamp: new Date('2025-02-03T09:15:00Z'),
    type: 'change',
    message: "Namespace 'ns-dev' lifecycle changed from active to deprecated",
  },
  {
    id: '34',
    timestamp: new Date('2025-02-04T09:21:00Z'),
    type: 'add',
    message: "Repository 'ns-dev/api' created",
  },
  {
    id: '35',
    timestamp: new Date('2025-02-04T10:07:00Z'),
    type: 'change',
    message: "Repository 'ns-dev/api' restricted to maintainers",
  },
  {
    id: '36',
    timestamp: new Date('2025-02-05T13:12:00Z'),
    type: 'add',
    message: "Tag 'v0.1.0' of 'ns-dev/api' marked as stable",
  },
  {
    id: '37',
    timestamp: new Date('2025-02-05T13:54:00Z'),
    type: 'change',
    message: "Tag 'v0.1.0' of 'ns-dev/api' marked as unstable",
  },
  {
    id: '38',
    timestamp: new Date('2025-02-06T08:11:00Z'),
    type: 'delete',
    message: "Maintainer removed from namespace 'ns-dev'",
  },
  {
    id: '39',
    timestamp: new Date('2025-02-07T11:27:00Z'),
    type: 'change',
    message: "Namespace 'ns-dev' lifecycle changed from deprecated to disabled",
  },
  {
    id: '40',
    timestamp: new Date('2025-02-08T16:40:00Z'),
    type: 'delete',
    message: "Repository 'ns-dev/api' was deleted",
  },
  {
    id: '41',
    timestamp: new Date('2025-02-09T09:30:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-dev'",
  },
  {
    id: '42',
    timestamp: new Date('2025-02-10T10:15:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' visibility changed from private to internal",
  },
  {
    id: '43',
    timestamp: new Date('2025-02-12T14:05:00Z'),
    type: 'add',
    message: "Repository 'ns-alpha/docs' created",
  },
  {
    id: '44',
    timestamp: new Date('2025-02-13T08:20:00Z'),
    type: 'add',
    message: "Tag 'v2.1.0' of 'ns-alpha/core' marked as stable",
  },
  {
    id: '45',
    timestamp: new Date('2025-02-14T11:11:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' lifecycle changed from disabled to archived",
  },
  {
    id: '46',
    timestamp: new Date('2025-02-16T09:17:00Z'),
    type: 'delete',
    message: "Repository 'ns-alpha/docs' was deleted",
  },
  {
    id: '47',
    timestamp: new Date('2025-02-17T07:50:00Z'),
    type: 'add',
    message: "Namespace 'ns-sec' created",
  },
  {
    id: '48',
    timestamp: new Date('2025-02-17T08:53:00Z'),
    type: 'add',
    message: "Maintainer assigned to namespace 'ns-sec'",
  },
  {
    id: '49',
    timestamp: new Date('2025-02-17T09:20:00Z'),
    type: 'add',
    message: "Repository 'ns-sec/scan' created",
  },
  {
    id: '50',
    timestamp: new Date('2025-02-18T14:40:00Z'),
    type: 'change',
    message: "Repository 'ns-sec/scan' restricted to maintainers",
  },

  {
    id: '51',
    timestamp: new Date('2025-02-19T13:12:00Z'),
    type: 'add',
    message: "Tag 'v0.5.0' of 'ns-sec/scan' marked as stable",
  },
  {
    id: '52',
    timestamp: new Date('2025-02-21T09:22:00Z'),
    type: 'change',
    message: "Tag 'v0.5.0' of 'ns-sec/scan' marked as deprecated",
  },
  {
    id: '53',
    timestamp: new Date('2025-02-23T15:40:00Z'),
    type: 'delete',
    message: "Maintainer removed from namespace 'ns-sec'",
  },
  {
    id: '54',
    timestamp: new Date('2025-02-25T10:10:00Z'),
    type: 'change',
    message: "Namespace 'ns-sec' lifecycle changed from active to deprecated",
  },
  {
    id: '55',
    timestamp: new Date('2025-02-27T12:14:00Z'),
    type: 'delete',
    message: "Repository 'ns-sec/scan' was deleted",
  },
  {
    id: '56',
    timestamp: new Date('2025-03-01T09:44:00Z'),
    type: 'add',
    message: "Namespace 'ns-tools' created",
  },
  {
    id: '57',
    timestamp: new Date('2025-03-01T10:16:00Z'),
    type: 'add',
    message: "Repository 'ns-tools/ci' created",
  },
  {
    id: '58',
    timestamp: new Date('2025-03-02T09:22:00Z'),
    type: 'add',
    message: "Developer added to namespace 'ns-tools'",
  },
  {
    id: '59',
    timestamp: new Date('2025-03-03T14:41:00Z'),
    type: 'change',
    message: "Repository 'ns-tools/ci' visibility changed from internal to public",
  },
  {
    id: '60',
    timestamp: new Date('2025-03-05T11:50:00Z'),
    type: 'add',
    message: "Tag 'v1.0.0' of 'ns-tools/ci' marked as stable",
  },

  {
    id: '61',
    timestamp: new Date('2025-03-07T09:55:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' visibility changed from internal to private",
  },
  {
    id: '62',
    timestamp: new Date('2025-03-08T12:10:00Z'),
    type: 'delete',
    message: "Tag 'v1.0.0' of 'ns-tools/ci' removed",
  },
  {
    id: '63',
    timestamp: new Date('2025-03-09T13:33:00Z'),
    type: 'add',
    message: "Tag 'v1.1.0' of 'ns-tools/ci' marked as stable",
  },
  {
    id: '64',
    timestamp: new Date('2025-03-10T15:20:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-tools'",
  },
  {
    id: '65',
    timestamp: new Date('2025-03-12T10:03:00Z'),
    type: 'change',
    message: "Repository 'ns-tools/ci' restricted to selected users",
  },
  {
    id: '66',
    timestamp: new Date('2025-03-14T08:22:00Z'),
    type: 'add',
    message: "Maintainer assigned to namespace 'ns-tools'",
  },
  {
    id: '67',
    timestamp: new Date('2025-03-15T11:40:00Z'),
    type: 'change',
    message: "Namespace 'ns-tools' lifecycle changed from active to deprecated",
  },
  {
    id: '68',
    timestamp: new Date('2025-03-16T09:50:00Z'),
    type: 'delete',
    message: "Guest removed from namespace 'ns-tools'",
  },
  {
    id: '69',
    timestamp: new Date('2025-03-17T09:47:00Z'),
    type: 'change',
    message: "Namespace 'ns-prod' visibility changed from public to private",
  },
  {
    id: '70',
    timestamp: new Date('2025-03-20T14:40:00Z'),
    type: 'add',
    message: "Repository 'ns-prod/web' created",
  },

  {
    id: '71',
    timestamp: new Date('2025-03-21T13:20:00Z'),
    type: 'add',
    message: "Tag 'v1.0.0' of 'ns-prod/web' marked as stable",
  },
  {
    id: '72',
    timestamp: new Date('2025-03-23T08:22:00Z'),
    type: 'change',
    message: "Tag 'v1.0.0' of 'ns-prod/web' marked as unstable",
  },
  {
    id: '73',
    timestamp: new Date('2025-03-25T10:42:00Z'),
    type: 'delete',
    message: "Repository 'ns-prod/web' was deleted",
  },
  {
    id: '74',
    timestamp: new Date('2025-03-27T07:15:00Z'),
    type: 'delete',
    message: "Developer removed from namespace 'ns-qa'",
  },
  {
    id: '75',
    timestamp: new Date('2025-03-29T11:10:00Z'),
    type: 'change',
    message: "Namespace 'ns-qa' lifecycle changed from public to deprecated",
  },
  {
    id: '76',
    timestamp: new Date('2025-04-02T10:05:00Z'),
    type: 'add',
    message: "Namespace 'ns-services' created",
  },
  {
    id: '77',
    timestamp: new Date('2025-04-02T10:33:00Z'),
    type: 'add',
    message: "Repository 'ns-services/auth' created",
  },
  {
    id: '78',
    timestamp: new Date('2025-04-03T08:27:00Z'),
    type: 'add',
    message: "Developer added to namespace 'ns-services'",
  },
  {
    id: '79',
    timestamp: new Date('2025-04-04T11:19:00Z'),
    type: 'add',
    message: "Tag 'v0.1.0' of 'ns-services/auth' marked as stable",
  },
  {
    id: '80',
    timestamp: new Date('2025-04-06T14:45:00Z'),
    type: 'change',
    message: "Namespace 'ns-services' visibility changed from internal to public",
  },

  {
    id: '81',
    timestamp: new Date('2025-04-08T12:33:00Z'),
    type: 'delete',
    message: "Tag 'v0.1.0' of 'ns-services/auth' removed",
  },
  {
    id: '82',
    timestamp: new Date('2025-04-10T10:42:00Z'),
    type: 'change',
    message: "Namespace 'ns-dev' visibility changed from disabled to internal",
  },
  {
    id: '83',
    timestamp: new Date('2025-04-12T08:40:00Z'),
    type: 'add',
    message: "Repository 'ns-dev/tester' created",
  },
  {
    id: '84',
    timestamp: new Date('2025-04-13T09:10:00Z'),
    type: 'add',
    message: "Tag 'v0.0.1' of 'ns-dev/tester' marked as unstable",
  },
  {
    id: '85',
    timestamp: new Date('2025-04-14T13:55:00Z'),
    type: 'change',
    message: "Repository 'ns-dev/tester' restricted to maintainers",
  },
  {
    id: '86',
    timestamp: new Date('2025-04-17T11:10:00Z'),
    type: 'delete',
    message: "Repository 'ns-dev/tester' was deleted",
  },
  {
    id: '87',
    timestamp: new Date('2025-04-19T09:22:00Z'),
    type: 'change',
    message: "Namespace 'ns-prod' lifecycle changed from private to deprecated",
  },
  {
    id: '88',
    timestamp: new Date('2025-04-21T08:44:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-dev'",
  },
  {
    id: '89',
    timestamp: new Date('2025-04-23T12:16:00Z'),
    type: 'add',
    message: "Repository 'ns-dev/lab' created",
  },
  {
    id: '90',
    timestamp: new Date('2025-04-25T15:14:00Z'),
    type: 'add',
    message: "Tag 'v2.0.0' of 'ns-dev/lab' marked as stable",
  },

  {
    id: '91',
    timestamp: new Date('2025-04-26T11:11:00Z'),
    type: 'change',
    message: "Tag 'v2.0.0' of 'ns-dev/lab' marked as deprecated",
  },
  {
    id: '92',
    timestamp: new Date('2025-04-28T10:42:00Z'),
    type: 'delete',
    message: "Repository 'ns-dev/lab' was deleted",
  },
  {
    id: '93',
    timestamp: new Date('2025-04-30T09:20:00Z'),
    type: 'change',
    message: "Namespace 'ns-dev' lifecycle changed from disabled to archived",
  },
  {
    id: '94',
    timestamp: new Date('2025-05-02T13:42:00Z'),
    type: 'add',
    message: "Namespace 'ns-ops' created",
  },
  {
    id: '95',
    timestamp: new Date('2025-05-03T11:14:00Z'),
    type: 'add',
    message: "Repository 'ns-ops/logs' created",
  },
  {
    id: '96',
    timestamp: new Date('2025-05-05T10:33:00Z'),
    type: 'change',
    message: "Repository 'ns-ops/logs' restricted to selected users",
  },
  {
    id: '97',
    timestamp: new Date('2025-05-06T12:05:00Z'),
    type: 'add',
    message: "Tag 'v0.0.2' of 'ns-ops/logs' marked as stable",
  },
  {
    id: '98',
    timestamp: new Date('2025-05-07T09:44:00Z'),
    type: 'delete',
    message: "Maintainer removed from namespace 'ns-ops'",
  },
  {
    id: '99',
    timestamp: new Date('2025-05-08T10:10:00Z'),
    type: 'change',
    message: "Namespace 'ns-ops' lifecycle changed from active to deprecated",
  },
  {
    id: '100',
    timestamp: new Date('2025-05-09T12:40:00Z'),
    type: 'delete',
    message: "Repository 'ns-ops/logs' was deleted",
  },

  // Continue monthly variations until end of year:

  {
    id: '101',
    timestamp: new Date('2025-06-02T09:44:00Z'),
    type: 'add',
    message: "Namespace 'ns-pay' created",
  },
  {
    id: '102',
    timestamp: new Date('2025-06-03T11:22:00Z'),
    type: 'add',
    message: "Repository 'ns-pay/gw' created",
  },
  {
    id: '103',
    timestamp: new Date('2025-06-04T10:33:00Z'),
    type: 'add',
    message: "Tag 'v1.0.0' of 'ns-pay/gw' marked as stable",
  },
  {
    id: '104',
    timestamp: new Date('2025-06-06T08:40:00Z'),
    type: 'change',
    message: "Repository 'ns-pay/gw' restricted to maintainers",
  },
  {
    id: '105',
    timestamp: new Date('2025-06-10T12:40:00Z'),
    type: 'change',
    message: "Namespace 'ns-pay' visibility changed from public to internal",
  },

  {
    id: '106',
    timestamp: new Date('2025-07-02T09:30:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-pay'",
  },
  {
    id: '107',
    timestamp: new Date('2025-07-04T11:50:00Z'),
    type: 'delete',
    message: "Tag 'v1.0.0' of 'ns-pay/gw' removed",
  },
  {
    id: '108',
    timestamp: new Date('2025-07-05T09:33:00Z'),
    type: 'change',
    message: "Namespace 'ns-pay' lifecycle changed from internal to deprecated",
  },

  {
    id: '109',
    timestamp: new Date('2025-08-08T10:12:00Z'),
    type: 'delete',
    message: "Repository 'ns-pay/gw' was deleted",
  },
  {
    id: '110',
    timestamp: new Date('2025-08-09T09:55:00Z'),
    type: 'add',
    message: "Namespace 'ns-arch' created",
  },

  {
    id: '111',
    timestamp: new Date('2025-09-02T08:33:00Z'),
    type: 'add',
    message: "Repository 'ns-arch/modules' created",
  },
  {
    id: '112',
    timestamp: new Date('2025-09-03T12:05:00Z'),
    type: 'add',
    message: "Tag 'v0.9.0' of 'ns-arch/modules' marked as stable",
  },
  {
    id: '113',
    timestamp: new Date('2025-09-05T15:14:00Z'),
    type: 'change',
    message: "Repository 'ns-arch/modules' restricted to maintainers",
  },

  {
    id: '114',
    timestamp: new Date('2025-10-01T09:22:00Z'),
    type: 'delete',
    message: "Developer removed from namespace 'ns-alpha'",
  },
  {
    id: '115',
    timestamp: new Date('2025-10-10T10:33:00Z'),
    type: 'change',
    message: "Namespace 'ns-alpha' lifecycle changed from archived to disabled",
  },
  {
    id: '116',
    timestamp: new Date('2025-10-15T08:42:00Z'),
    type: 'add',
    message: "Repository 'ns-alpha/qa-tools' created",
  },
  {
    id: '117',
    timestamp: new Date('2025-10-18T11:20:00Z'),
    type: 'add',
    message: "Tag 'v1.0.1' of 'ns-alpha/qa-tools' marked as unstable",
  },

  {
    id: '118',
    timestamp: new Date('2025-10-20T14:16:00Z'),
    type: 'add',
    message: "Guest added to namespace 'ns-arch'",
  },
  {
    id: '119',
    timestamp: new Date('2025-10-24T09:00:00Z'),
    type: 'change',
    message: "Namespace 'ns-arch' visibility changed from private to public",
  },
  {
    id: '120',
    timestamp: new Date('2025-10-29T10:30:00Z'),
    type: 'add',
    message: "Repository 'ns-arch/framework' created",
  },
];

const mockNamespaceRepositories: NamespaceRepositoryInfo[] = [
  {
    id: 'repo-001',
    name: 'webapp-backend',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-01-12T09:22:00Z'),
    created_by: 'alice',
    tags_count: '12',
    state: 'active',
  },
  {
    id: 'repo-002',
    name: 'webapp-frontend',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-01-14T13:10:00Z'),
    created_by: 'alice',
    tags_count: '8',
    state: 'active',
  },
  {
    id: 'repo-003',
    name: 'auth-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-01-18T15:40:00Z'),
    created_by: 'bob',
    tags_count: '5',
    state: 'active',
  },
  {
    id: 'repo-004',
    name: 'notification-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-02-01T09:00:00Z'),
    created_by: 'carol',
    tags_count: '17',
    state: 'active',
  },
  {
    id: 'repo-005',
    name: 'payment-gateway',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-02-03T13:55:00Z'),
    created_by: 'carol',
    tags_count: '9',
    state: 'deprecated',
  },
  {
    id: 'repo-006',
    name: 'analytics-engine',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-02-07T08:25:00Z'),
    created_by: 'dave',
    tags_count: '14',
    state: 'active',
  },
  {
    id: 'repo-007',
    name: 'data-warehouse',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-03-01T11:29:00Z'),
    created_by: 'emma',
    tags_count: '6',
    state: 'active',
  },
  {
    id: 'repo-008',
    name: 'cache-proxy',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-03-05T16:46:00Z'),
    created_by: 'emma',
    tags_count: '3',
    state: 'disabled',
  },
  {
    id: 'repo-009',
    name: 'ingress-controller',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-03-10T08:14:00Z'),
    created_by: 'frank',
    tags_count: '11',
    state: 'active',
  },
  {
    id: 'repo-010',
    name: 'event-bus',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-03-18T07:21:00Z'),
    created_by: 'frank',
    tags_count: '10',
    state: 'deprecated',
  },
  {
    id: 'repo-011',
    name: 'messaging-queue',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-04-01T09:50:00Z'),
    created_by: 'alice',
    tags_count: '4',
    state: 'active',
  },
  {
    id: 'repo-012',
    name: 'document-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-04-04T12:12:00Z'),
    created_by: 'bob',
    tags_count: '13',
    state: 'active',
  },
  {
    id: 'repo-013',
    name: 'file-storage',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-04-10T14:38:00Z'),
    created_by: 'carol',
    tags_count: '5',
    state: 'active',
  },
  {
    id: 'repo-014',
    name: 'cdn-edge',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-04-14T17:00:00Z'),
    created_by: 'carol',
    tags_count: '7',
    state: 'disabled',
  },
  {
    id: 'repo-015',
    name: 'user-profile',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-05-02T10:40:00Z'),
    created_by: 'emma',
    tags_count: '15',
    state: 'active',
  },
  {
    id: 'repo-016',
    name: 'preference-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-05-03T11:11:00Z'),
    created_by: 'emma',
    tags_count: '6',
    state: 'deprecated',
  },
  {
    id: 'repo-017',
    name: 'audit-logger',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-05-08T09:42:00Z'),
    created_by: 'dave',
    tags_count: '4',
    state: 'active',
  },
  {
    id: 'repo-018',
    name: 'session-manager',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-05-14T15:28:00Z'),
    created_by: 'dave',
    tags_count: '10',
    state: 'active',
  },
  {
    id: 'repo-019',
    name: 'proxy-router',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-05-19T18:20:00Z'),
    created_by: 'frank',
    tags_count: '3',
    state: 'disabled',
  },
  {
    id: 'repo-020',
    name: 'log-forwarder',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-06-01T07:00:00Z'),
    created_by: 'bob',
    tags_count: '9',
    state: 'active',
  },
  {
    id: 'repo-021',
    name: 'helm-charts',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-06-07T16:30:00Z'),
    created_by: 'carol',
    tags_count: '20',
    state: 'active',
  },
  {
    id: 'repo-022',
    name: 'terraform-modules',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-06-09T15:40:00Z'),
    created_by: 'carol',
    tags_count: '11',
    state: 'active',
  },
  {
    id: 'repo-023',
    name: 'nginx-custom',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-06-13T12:11:00Z'),
    created_by: 'dave',
    tags_count: '2',
    state: 'deprecated',
  },
  {
    id: 'repo-024',
    name: 'reverse-proxy',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-06-15T09:55:00Z'),
    created_by: 'dave',
    tags_count: '8',
    state: 'active',
  },
  {
    id: 'repo-025',
    name: 'gateway-api',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-07-02T10:44:00Z'),
    created_by: 'emma',
    tags_count: '12',
    state: 'active',
  },
  {
    id: 'repo-026',
    name: 'device-manager',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-07-04T14:15:00Z'),
    created_by: 'bob',
    tags_count: '4',
    state: 'disabled',
  },
  {
    id: 'repo-027',
    name: 'firmware-updater',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-07-09T09:32:00Z'),
    created_by: 'carol',
    tags_count: '5',
    state: 'active',
  },
  {
    id: 'repo-028',
    name: 'edge-collector',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-07-11T08:28:00Z'),
    created_by: 'carol',
    tags_count: '6',
    state: 'active',
  },
  {
    id: 'repo-029',
    name: 'edge-cache',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-07-14T11:45:00Z'),
    created_by: 'frank',
    tags_count: '8',
    state: 'deprecated',
  },
  {
    id: 'repo-030',
    name: 'monitoring-agent',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-08-05T06:57:00Z'),
    created_by: 'alice',
    tags_count: '16',
    state: 'active',
  },
  {
    id: 'repo-031',
    name: 'alert-engine',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-08-06T08:30:00Z'),
    created_by: 'emma',
    tags_count: '13',
    state: 'active',
  },
  {
    id: 'repo-032',
    name: 'metrics-scraper',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-08-09T17:23:00Z'),
    created_by: 'emma',
    tags_count: '7',
    state: 'active',
  },
  {
    id: 'repo-033',
    name: 'event-correlator',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-08-10T13:08:00Z'),
    created_by: 'bob',
    tags_count: '4',
    state: 'deprecated',
  },
  {
    id: 'repo-034',
    name: 'billing-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-01T11:32:00Z'),
    created_by: 'carol',
    tags_count: '15',
    state: 'active',
  },
  {
    id: 'repo-035',
    name: 'invoice-generator',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-04T10:21:00Z'),
    created_by: 'carol',
    tags_count: '8',
    state: 'active',
  },
  {
    id: 'repo-036',
    name: 'tax-compute',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-05T08:48:00Z'),
    created_by: 'bob',
    tags_count: '5',
    state: 'disabled',
  },
  {
    id: 'repo-037',
    name: 'currency-sync',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-06T12:16:00Z'),
    created_by: 'emma',
    tags_count: '6',
    state: 'active',
  },
  {
    id: 'repo-038',
    name: 'exchange-rates',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-10T07:29:00Z'),
    created_by: 'alice',
    tags_count: '9',
    state: 'active',
  },
  {
    id: 'repo-039',
    name: 'wallet-service',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-12T10:43:00Z'),
    created_by: 'dave',
    tags_count: '3',
    state: 'deprecated',
  },
  {
    id: 'repo-040',
    name: 'payment-collector',
    namesapce: 'alpha',
    namespace_id: 'ns-001',
    created_at: new Date('2024-09-15T12:55:00Z'),
    created_by: 'frank',
    tags_count: '11',
    state: 'active',
  },
];

const NamespaceViewPage = () => {
  const [userAccessList, setUserAccessList] = useState<UserAccessInfo[]>(namespaceMaintainers);
  const [repositoryList, setRepositoryList] =
    useState<NamespaceRepositoryInfo[]>(mockNamespaceRepositories);

  const navigate = useNavigate();

  const handlePeriodChange = (start: Date, end: Date): void => { };

  const handleUserAccessFilter = (options: string[]): void => { };

  const navigateToRepository = (nsId: string, id: string): void => {
    navigate(`/console/access-management/repositories/${id}`);
  };

  return (
    <div className="flex-grow-1">
      <div className="bg-white border-round-lg flex flex-column gap-2">
        <div className="flex flex-column pt-2 gap-2">
          <div className="flex flex-row justify-content-between pt-2 pb-2">
            <div className="pl-3">
              <span className="font-semibold text-lg">Namespace:</span>
              &nbsp;&nbsp;&nbsp;
              <span className=" font-medium text-lg">team-platform-eng</span>
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
            </div>
          </div>
          <Divider layout="horizontal" className="mt-0 pt-0" />

          <div className="flex flex-row justify-content-between pl-3 pr-3 pb-1">
            <div className="flex flex-row align-items-center">
              <span className="text-sm">Activity (Last 12 Months)</span>
              &nbsp;&nbsp;
              <span className="pi pi-info-circle text-sm"></span>
            </div>

            <div className="pr-8 text-xs text-color-secondary flex flex-row gap-2">
              <div>
                <span className="legend legend-add"></span>
                &nbsp;&nbsp;
                <span>Additions</span>
              </div>
              <div>
                <span className="legend legend-change"></span>
                &nbsp;&nbsp;
                <span>Changes</span>
              </div>
              <div>
                <span className="legend legend-delete"></span>
                &nbsp;&nbsp;
                <span>Deletions</span>
              </div>
            </div>
          </div>

          <div className="pl-3">
            <ChangeTrackerView
              width={1000}
              height={450}
              data={mockEvents}
              filters={{ add: true, delete: true, change: true }}
              period={12}
              onPeriodChange={handlePeriodChange}
            />
          </div>

          <div className="flex flex-row  gap-3">
            {/* <div className='font-medium'>Change History</div> */}
          </div>

          <div>
            <TabView>
              <TabPanel header={<div className="font-medium">User Access</div>}>
                <div className="flex flex-row align-items-center justify-content-between pl-3 pr-2 ">
                  <div className="flex align-items-center  flex-grow-1">
                    <div className=" text-sm"> Select Filter</div>
                    <ChipsFilter
                      filterOptions={USER_ACCESS_FILTER_OPTIONS}
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
                  value={userAccessList}
                  paginator
                  paginatorLeft={<div className="text-xs">82 users found</div>}
                  rows={8}
                  totalRecords={namespaceMaintainers.length}
                >
                  <Column field="username" header="User" sortable />
                  <Column field="access_level" header="Access Level" sortable />

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
              <TabPanel header={<div className="font-medium">Repositories</div>}>
                <div className="flex flex-row align-items-center justify-content-between pl-3 pr-2 ">
                  <div className="flex align-items-center  flex-grow-1">
                    <div className=" text-sm"> Select Filter</div>
                    <ChipsFilter
                      filterOptions={REPOSITORY_FILTER_OPTIONS}
                      maxChipsPerRow={4}
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
                      placeholder="Search repositories . . . ."
                    />
                    <i className="pi pi-search text-teal-400 text-sm pr-2 "></i>
                  </Button>
                </div>
                <DataTable
                  value={repositoryList}
                  paginator
                  paginatorLeft={<div className="text-xs">82 repositories found</div>}
                  rows={8}
                  totalRecords={repositoryList.length}
                >
                  <Column field="name" header="Name" sortable />
                  <Column field="tags_count" header="Tags" align="center" sortable />
                  {/* If namespace state is not active, we should display the column header as "Effective State" and tooltip to explain the inheritance*/}
                  <Column
                    header="Effective State"
                    align="center"
                    body={(data: NamespaceRepositoryInfo) => {
                      return <span className={`status-circle status-${data.state}`}></span>;
                    }}
                  />

                  <Column header="Created By" field="created_by" sortable />
                  <Column
                    header="Created At"
                    field="created_at"
                    body={(data: NamespaceRepositoryInfo) => data.created_at.toUTCString()}
                    sortable
                  />
                  <Column
                    align="right"
                    header={
                      <div>
                        <Button
                          size="small"
                          className="p-2 m-0 border-1 border-solid border-round-lg border-teal-100 text-xs"
                        // onClick={() => setShowCreateUserAccountDialog((c) => !c)}
                        >
                          <span className="pi pi-plus text-xs"></span>
                          &nbsp;&nbsp;
                          <span>Create Repository</span>
                        </Button>
                      </div>
                    }
                    body={(data: NamespaceRepositoryInfo) => {
                      return (
                        <div className="flex flex-row align-items-center gap-2 justify-content-end">
                          <Button
                            className="border-round-3xl border-1 text-sm"
                            outlined
                            size="small"
                            onClick={() => navigateToRepository(data.namespace_id, data.id)}
                          >
                            <span className="pi pi-wrench text-sm"></span>
                            &nbsp;&nbsp;
                            <span className="text-sm">Access</span>
                          </Button>

                          {/* <span className='pi pi-ellipsis-v text-sm'></span> */}
                        </div>
                      );
                    }}
                  />

                  {/* <Column body={(ns: NamespaceRepositoryInfo) => <Button
                  className="border-round-3xl border-1 text-sm"
                  outlined
                  size="small"
                // onClick={() => handleManageAccess(ns.id)}
                >
                  <span className="pi pi-wrench text-sm"></span>
                  &nbsp;&nbsp;
                  <span className='text-sm'>Access</span>
                </Button>} /> */}
                </DataTable>
              </TabPanel>
            </TabView>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NamespaceViewPage;
