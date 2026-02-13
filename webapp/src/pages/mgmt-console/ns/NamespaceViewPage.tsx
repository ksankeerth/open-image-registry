import { Button } from 'primereact/button';
import { Chip } from 'primereact/chip';
import { TabPanel, TabView } from 'primereact/tabview';
import React, { useCallback, useEffect, useState } from 'react';
import { ChangeTrackerEventInfo, NamespaceRepositoryInfo } from '../../../types/app_types';
import ChangeTrackerView from '../../../components/ChangeTrackerView';
import { Divider } from 'primereact/divider';
import ChipsFilter from '../../../components/ChipsFilter';
import {
  REPOSITORY_FILTER_OPTIONS,
  USER_ACCESS_FILTER_OPTIONS,
} from '../../../config/table_filter';
import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { useNavigate, useParams } from 'react-router-dom';
import SearchButton from '../../../components/SearchButton';
import { useToast } from '../../../components/ToastComponent';
import {
  getNamespaceUsers,
  getResourceNamespacesById,
  getResourceRepositories,
  GetResourceRepositoriesData,
  NamespaceResponse,
  RepositoryViewDto,
  ResourceAccessViewDto,
} from '../../../api';
import { useLoader } from '../../../components/loader';
import { classNames } from 'primereact/utils';
import CreateRepositoryDialog from '../../../components/CreateRepository';
import { Tooltip } from 'primereact/tooltip';

// eslint-disable-next-line react-refresh/only-export-components
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

const NameColumnTemplate = (props: { r: RepositoryViewDto; n: NamespaceResponse }) => {
  const toolTipTarget = `rs-${props.r.name}-repo-stats`;

  return (
    <React.Fragment>
      <div className={`flex flex-row align-items-center gap-2 cursor-pointer ${toolTipTarget}`}>
        {props.n.state !== 'Active' && (
          <>
            {props.r.state == 'Disabled' && <div className="status-circle status-disabled" />}
            {props.r.state == 'Deprecated' && <div className="status-circle status-deprecated" />}
            {/* TODO:, We have to tell about effective state */}
            <span className="pi pi-info-circle text-sm"></span>
          </>
        )}
        {props.n.state === 'Active' && (
          <>
            {props.r.state == 'Active' && <div className="status-circle status-active" />}
            {props.r.state == 'Disabled' && <div className="status-circle status-disabled" />}
            {props.r.state == 'Deprecated' && <div className="status-circle status-deprecated" />}
          </>
        )}

        <div className="font-inter font-semibold">{props.r.name}</div>
        {props.r.is_public && (
          <Chip
            className="pl-2 pr-2 p-0  bg-white border-600 border-1 font-bold"
            style={{ fontSize: '0.60rem' }}
            label="public"
          />
        )}
      </div>
      <Tooltip target={'.' + toolTipTarget}>
        <div className="">TODO: Show Repository Stats</div>
      </Tooltip>
    </React.Fragment>
  );
};

const NamespaceViewPage = () => {
  const { id } = useParams();

  const { showError } = useToast();

  const { showLoading, hideLoading } = useLoader();

  const [nsInfo, setNsInfo] = useState<NamespaceResponse>();

  const [userAccessList, setUserAccessList] = useState<ResourceAccessViewDto[]>([]);
  const [repositoryList, setRepositoryList] = useState<RepositoryViewDto[]>([]);
  const [totalUsers, setTotalUsers] = useState<number>(0);
  const [totalRepos, setTotalRepos] = useState<number>(0);

  const [showCreateRepoDialog, setShowCreateRepoDialog] = useState<boolean>(false);

  const navigate = useNavigate();

  const handlePeriodChange = (start: Date, end: Date): void => { };

  const handleUserAccessFilter = (options: string[]): void => { };

  const navigateToRepository = (repoId: string): void => {
    navigate(`/console/access-management/repositories/${repoId}`);
  };

  const loadNamespace = useCallback(async () => {
    if (!id) {
      return;
    }

    showLoading('Loading namespace ...');

    const { data, error } = await getResourceNamespacesById({
      path: {
        id: id,
      },
    });

    hideLoading();

    if (error) {
      showError(error.error_message);
    } else {
      if (data) {
        setTimeout(() => {
          setNsInfo(data);
        }, 0);
      }
    }
  }, [id, showLoading, hideLoading, showError]);

  useEffect(() => {
    loadNamespace();
  }, [loadNamespace]);

  const loadUsers = useCallback(async () => {
    if (!id) {
      return;
    }

    showLoading('Loading users ...');
    const { data, error } = await getNamespaceUsers({
      path: {
        id: id,
      },
    });
    hideLoading();

    if (error) {
      showError('Failed to load user access!');
    } else {
      setTimeout(() => {
        setTotalUsers(data.total);
        if (data.entities) {
          for (let i = 0; i < data.total - data.entities.length; i++) {
            data.entities.push({} as ResourceAccessViewDto); // adding fake data for showing pagination properly
          }
          setUserAccessList(data.entities);
        }
      }, 0);
    }
  }, [hideLoading, id, showError, showLoading]);

  const loadRepositories = useCallback(async () => {
    if (!id) {
      return;
    }
    showLoading('Loading users ...');

    const req: GetResourceRepositoriesData = {
      url: '/resource/repositories',
      query: {
        namespace_id: id,
      },
    };

    if (!id) {
      return;
    }

    const { data, error } = await getResourceRepositories(req);
    hideLoading();

    if (error) {
      showError('Failed to load repositories!');
    } else {
      setTimeout(() => {
        setTotalRepos(data.total);
        if (data.entities) {
          for (let i = 0; i < data.total - data.entities.length; i++) {
            data.entities.push({} as RepositoryViewDto); // adding fake data for showing pagination properly
          }
          setRepositoryList(data.entities);
        }
      }, 0);
    }
  }, [hideLoading, id, showError, showLoading]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  useEffect(() => {
    loadRepositories();
  }, [loadRepositories]);

  return (
    <div className="flex-grow-1 gap-3">
      <div className="flex flex-column gap-2 my-2">
        <div className="flex flex-column gap-3">
          {/* Title */}
          <div className="flex flex-row align-items-stretch justify-content-between max-h-4rem h-3rem">
            <div className="pl-2 flex w-full justify-content-between align-items-center">
              <div className="flex flex-grow-0  align-items-center max-h-4rem h-3rem h-full">
                <span className="text-2xl title-500">
                  Namespace:
                  {nsInfo ? '  ' + nsInfo.name : 'Loading ...'}
                </span>
                &nbsp;&nbsp;
                <span className="h-full flex align-items-start gap-2">
                  {nsInfo?.state && (
                    <Chip
                      className={classNames(
                        'pl-2 pr-2 p-0  bg-white border-600 border-1 ',
                        nsInfo.state == 'Active'
                          ? 'border-teal-300'
                          : nsInfo.state == 'Deprecated'
                            ? 'border-orange-300'
                            : 'border-red-300'
                      )}
                      style={{ fontSize: '0.60rem' }}
                      label={nsInfo.state}
                    />
                  )}

                  <Chip
                    className="pl-2 pr-2 p-0  bg-white border-600 border-1"
                    style={{ fontSize: '0.60rem' }}
                    label={nsInfo?.is_public ? 'Public' : 'Private'}
                  />
                </span>
              </div>
              <div className="flex-grow-0">
                <div>
                  <span className="pi pi-ellipsis-v text-sm text-color"></span>
                </div>
              </div>
            </div>
          </div>

          {/* Activity */}
          <div className="flex flex-column gap-0 pr-0 mb-0 pb-1  mt-2 border-1 border-100 shadow-micro bg-white border-round-xl">
            <div className="flex flex-row justify-content-between align-items-stretch px-3 h-full">
              <div className="flex flex-row align-items-center h-full">
                <span className="text-md font-medium title-400 py-4">
                  Activity (Last 12 Months)
                </span>
                &nbsp;&nbsp;
                <span className="pi pi-info-circle text-sm"></span>
              </div>
              <div className="flex-grow-1 flex justify-content-end">
                <div className="flex h-full align-items-end gap-4 pb-2">
                  <div className="flex align-items-end gap-2">
                    <span className="pi pi-user text-xs text-600"></span>
                    <span className="text-xs text-600">{nsInfo?.created_by}</span>
                  </div>
                  <div className="flex align-items-end gap-2">
                    <span className="pi pi-calendar-clock text-xs text-600"></span>
                    <span className="text-xs text-600">{nsInfo?.created_at}</span>
                  </div>
                </div>
              </div>
            </div>
            <Divider layout="horizontal" className="p-0 m-0 mb-3" />
            <div className="flex flex-row justify-content-start pl-2 pb-0 pt-0">
              <div className="text-xs text-color-secondary flex flex-row px-3 gap-6">
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
            <div className="flex flex-column">
              <ChangeTrackerView
                data={mockEvents}
                filters={{ add: true, delete: true, change: true }}
                period={12}
                onPeriodChange={handlePeriodChange}
              />
            </div>
          </div>

          <div className="mt-2 pt-2 pb-1 border-1 border-100 bg-white border-round-xl shadow-micro">
            <TabView className="border-round-lg">
              <TabPanel header={<div className="font-medium">User Access</div>}>
                <div className="flex flex-row align-items-center justify-content-between pl-3 pr-2">
                  <div className="flex align-items-center  flex-grow-1">
                    <div className="text-sm">Select Filter</div>
                    <ChipsFilter
                      filterOptions={USER_ACCESS_FILTER_OPTIONS}
                      handleFilterChange={handleUserAccessFilter}
                    />
                  </div>
                  <SearchButton
                    placeholder="Search users ...."
                    class="shadow-none  border-1 border-100 hover:shadow-micro hover:border-teal-100"
                    handleSearch={() => { }}
                  />
                </div>
                <DataTable
                  value={userAccessList}
                  paginator
                  paginatorLeft={<div className="text-xs">{totalUsers} users found</div>}
                  rows={8}
                  totalRecords={totalUsers}
                  className="pb-0"
                  pt={{
                    paginator: {
                      root: {
                        className: 'border-none pb-0 mb-0',
                      },
                    },
                  }}
                >
                  <Column field="username" header="User" sortable />
                  <Column field="access_level" header="Access Level" sortable />

                  <Column header="Granted By" field="granted_by" sortable />
                  <Column
                    header="Granted At"
                    field="granted_at"
                    body={(data: ResourceAccessViewDto) => data.granted_at}
                    sortable
                  />
                  <Column
                    align="right"
                    header={
                      <div>
                        <Button
                          size="large"
                          severity="warning"
                          className="border-1 pl-3 pr-3   border-solid border-round-3xl border-teal-100 text-xs shadow-micro"
                        >
                          <span className="pi pi-shield text-xs"></span>
                          &nbsp;&nbsp;&nbsp;
                          <span className="font-inter text-sm">Grant Access</span>
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
                  <div className="flex align-items-center  flex-grow-1 bg-white">
                    <div className=" text-sm"> Select Filter</div>
                    <ChipsFilter
                      filterOptions={REPOSITORY_FILTER_OPTIONS}
                      maxChipsPerRow={4}
                      handleFilterChange={handleUserAccessFilter}
                    />
                  </div>
                  <SearchButton
                    placeholder="Search repositories ...."
                    class="shadow-none  border-1 border-100 hover:shadow-micro hover:border-teal-100"
                    handleSearch={() => { }}
                  />
                </div>
                <DataTable
                  value={repositoryList}
                  paginator
                  paginatorLeft={<div className="text-xs">{totalRepos} repositories found</div>}
                  rows={8}
                  totalRecords={repositoryList.length}
                >
                  <Column
                    field="name"
                    header="Name"
                    sortable
                    body={(data: RepositoryViewDto) => {
                      return <NameColumnTemplate r={data} n={nsInfo as NamespaceResponse} />;
                    }}
                  />
                  <Column field="tags" header="Tags" align="center" sortable />
                  <Column header="Created By" field="created_by" sortable />
                  <Column
                    header="Created At"
                    field="created_at"
                    body={(data: RepositoryViewDto) => data.created_at}
                    sortable
                  />
                  <Column
                    align="right"
                    header={
                      <div>
                        <Button
                          size="large"
                          className="border-1 pl-3 pr-3   border-solid border-round-3xl border-teal-100 text-xs shadow-micro"
                          onClick={() => setShowCreateRepoDialog(true)}
                        >
                          <span className="pi pi-plus text-xs"></span>
                          &nbsp;&nbsp;&nbsp;
                          <span className="font-inter text-sm">Create Repository</span>
                        </Button>
                      </div>
                    }
                    body={(data: RepositoryViewDto) => {
                      return (
                        <div className="flex flex-row align-items-center gap-2 justify-content-end">
                          <Button
                            className="border-round-3xl border-1 text-sm"
                            outlined
                            size="small"
                            onClick={() => navigateToRepository(data.id as string)}
                          >
                            <span className="pi pi-eye text-sm"></span>
                            &nbsp;&nbsp;
                            <span className="text-sm">View</span>
                          </Button>
                          {/* <Button
                            className="border-round-3xl border-1 text-sm"
                            outlined
                            size="small"
                            onClick={() => navigateToRepository(data.namespace_id, data.id)}
                          >
                            <span className="pi pi-wrench text-sm"></span>
                            &nbsp;&nbsp;
                            <span className="text-sm">Access</span>
                          </Button> */}

                          {/* <span className='pi pi-ellipsis-v text-sm'></span> */}
                        </div>
                      );
                    }}
                  />
                </DataTable>
                <CreateRepositoryDialog
                  namespaceId={id as string}
                  visible={showCreateRepoDialog}
                  hideCallback={() => setShowCreateRepoDialog(false)}
                />
              </TabPanel>
            </TabView>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NamespaceViewPage;
