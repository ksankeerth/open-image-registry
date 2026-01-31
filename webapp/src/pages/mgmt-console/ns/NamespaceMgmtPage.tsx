import React, { useCallback, useRef, useState } from 'react';

import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { Button } from 'primereact/button';

import { Chip } from 'primereact/chip';
import { OverlayPanel } from 'primereact/overlaypanel';
import { Tooltip } from 'primereact/tooltip';
import { useNavigate } from 'react-router-dom';
import ChipsFilter from '../../../components/ChipsFilter';
import CreateNamespaceDialog from '../../../components/CreateNamespace';
import SearchButton from '../../../components/SearchButton';
import { NAMESPACE_FILTER_OPTIONS } from '../../../config/table_filter';
import { NamespaceInfo, TableFilterSearchPaginationSortState } from '../../../types/app_types';
import { getRegistryNamespaces, GetRegistryNamespacesData, NamespaceViewDto } from '../../../api';
import { useLoader } from '../../../components/loader';

const mockNamespaces: NamespaceInfo[] = [
  {
    id: 'ns_001',
    name: 'openimage-core',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Core components and base images for OpenImageRegistry.',
    created_at: new Date('2024-02-15T10:12:00Z'),
    created_by: 'admin',
    updated_at: new Date('2024-03-10T12:00:00Z'),
    user_access: { maintainers: 3, developers: 8, guests: 5 },
    repositories: { total: 25, active: 22, deprecated: 2, disabled: 1, restricted: 4 },
  },
  {
    id: 'ns_002',
    name: 'frontend-ui',
    is_public: false,
    type: 'team',
    state: 'active',
    description: 'Frontend and UI image build pipelines.',
    created_at: new Date('2023-12-01T08:00:00Z'),
    created_by: 'alice',
    user_access: { maintainers: 2, developers: 6, guests: 0 },
    repositories: { total: 12, active: 12, deprecated: 0, disabled: 0, restricted: 1 },
  },
  {
    id: 'ns_003',
    name: 'backend-services',
    is_public: false,
    type: 'team',
    state: 'active',
    description: 'Microservices and backend APIs.',
    created_at: new Date('2024-01-10T14:22:00Z'),
    created_by: 'bob',
    user_access: { maintainers: 4, developers: 10, guests: 2 },
    repositories: { total: 30, active: 28, deprecated: 1, disabled: 1, restricted: 6 },
  },
  {
    id: 'ns_004',
    name: 'analytics',
    is_public: false,
    type: 'project',
    state: 'deprecated',
    description: 'Old analytics pipeline (deprecated).',
    created_at: new Date('2022-05-14T09:00:00Z'),
    created_by: 'carol',
    updated_at: new Date('2024-01-20T17:40:00Z'),
    user_access: { maintainers: 1, developers: 3, guests: 0 },
    repositories: { total: 10, active: 0, deprecated: 9, disabled: 1, restricted: 0 },
  },
  {
    id: 'ns_005',
    name: 'devtools',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Developer tooling images.',
    created_at: new Date('2023-06-18T11:45:00Z'),
    created_by: 'daniel',
    user_access: { maintainers: 2, developers: 12, guests: 7 },
    repositories: { total: 18, active: 17, deprecated: 1, disabled: 0, restricted: 2 },
  },
  {
    id: 'ns_006',
    name: 'python-libs',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Python runtime and base images.',
    created_at: new Date('2024-03-05T15:33:00Z'),
    created_by: 'eve',
    user_access: { maintainers: 1, developers: 4, guests: 10 },
    repositories: { total: 9, active: 9, deprecated: 0, disabled: 0, restricted: 0 },
  },
  {
    id: 'ns_007',
    name: 'node-libs',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Node.js base images, runtime variants.',
    created_at: new Date('2023-10-04T07:12:00Z'),
    created_by: 'admin',
    user_access: { maintainers: 1, developers: 3, guests: 12 },
    repositories: { total: 14, active: 14, deprecated: 0, disabled: 0, restricted: 0 },
  },
  {
    id: 'ns_008',
    name: 'security',
    is_public: false,
    type: 'team',
    state: 'active',
    description: 'Internal security scanning tools and hardened images.',
    created_at: new Date('2023-09-18T10:00:00Z'),
    created_by: 'bob',
    user_access: { maintainers: 3, developers: 5, guests: 0 },
    repositories: { total: 7, active: 6, deprecated: 1, disabled: 0, restricted: 5 },
  },
  {
    id: 'ns_009',
    name: 'openimage-examples',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Example images for demos and tutorials.',
    created_at: new Date('2024-04-01T09:55:00Z'),
    created_by: 'admin',
    user_access: { maintainers: 1, developers: 1, guests: 50 },
    repositories: { total: 5, active: 5, deprecated: 0, disabled: 0, restricted: 0 },
  },
  {
    id: 'ns_010',
    name: 'legacy-services',
    is_public: false,
    type: 'project',
    state: 'disabled',
    description: 'Legacy services pending removal.',
    created_at: new Date('2021-02-02T11:00:00Z'),
    created_by: 'frank',
    user_access: { maintainers: 1, developers: 0, guests: 0 },
    repositories: { total: 6, active: 0, deprecated: 3, disabled: 3, restricted: 0 },
  },
  {
    id: 'ns_011',
    name: 'mobile-app',
    is_public: false,
    type: 'team',
    state: 'active',
    description: 'Mobile application backend and UI assets.',
    created_at: new Date('2024-02-10T08:10:00Z'),
    created_by: 'eva',
    user_access: { maintainers: 2, developers: 8, guests: 3 },
    repositories: { total: 11, active: 10, deprecated: 1, disabled: 0, restricted: 1 },
  },
  {
    id: 'ns_012',
    name: 'ai-ml',
    is_public: false,
    type: 'project',
    state: 'active',
    description: 'Images for machine learning workflows.',
    created_at: new Date('2023-11-11T11:11:00Z'),
    created_by: 'alice',
    user_access: { maintainers: 3, developers: 15, guests: 5 },
    repositories: { total: 20, active: 18, deprecated: 1, disabled: 1, restricted: 4 },
  },
  {
    id: 'ns_013',
    name: 'go-runtime',
    is_public: true,
    type: 'project',
    state: 'active',
    description: 'Golang build and runtime base images.',
    created_at: new Date('2024-01-01T00:00:00Z'),
    created_by: 'admin',
    user_access: { maintainers: 1, developers: 3, guests: 30 },
    repositories: { total: 10, active: 10, deprecated: 0, disabled: 0, restricted: 0 },
  },
  {
    id: 'ns_014',
    name: 'proxy-cache',
    is_public: false,
    type: 'project',
    state: 'active',
    description: 'Cached images from external registries.',
    created_at: new Date('2024-02-22T16:05:00Z'),
    created_by: 'system',
    user_access: { maintainers: 2, developers: 0, guests: 0 },
    repositories: { total: 100, active: 100, deprecated: 0, disabled: 0, restricted: 100 },
  },
  {
    id: 'ns_015',
    name: 'qa-testing',
    is_public: false,
    type: 'team',
    state: 'active',
    description: 'Temporary images used for QA testing.',
    created_at: new Date('2024-03-17T09:30:00Z'),
    created_by: 'dave',
    user_access: { maintainers: 1, developers: 12, guests: 2 },
    repositories: { total: 40, active: 35, deprecated: 3, disabled: 2, restricted: 5 },
  },
];

const NameColumnTemplate = (ns: NamespaceInfo) => {
  const toolTipTarget = `ns-${ns.name}-repo-stats`;

  return (
    <React.Fragment>
      <div className={`flex flex-row align-items-center gap-2 cursor-pointer ${toolTipTarget}`}>
        {ns.state == 'active' && <div className="status-circle status-active" />}
        {ns.state == 'disabled' && <div className="status-circle status-disabled" />}
        {ns.state == 'deprecated' && <div className="status-circle status-deprecated" />}
        <div className="font-inter font-semibold">{ns.name}</div>
        {ns.is_public && (
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

const NamespaceMgmtPage = () => {
  const [showCreateNamespaceDialog, setShowCreateNamespaceDialog] = useState<boolean>(false);

  const [filterSortPagSearchState, setFilterSortPagSearchState] = useState<
    TableFilterSearchPaginationSortState<NamespaceViewDto>
  >({
    pagination: {
      page: 1,
      limit: 8,
    },
    sort: {
      key: 'name',
      order: 1,
    },
  });

  const navigate = useNavigate();

  const [talbeData, setTableData] = useState<NamespaceViewDto[]>([]);
  const [totalEntries, setTotalEntries] = useState<number>(0);

  const { showError } = useToast();

  const { showLoading, hideLoading } = useLoader();

  const loadNamespaces = useCallback(async () => {
    showLoading('Loading namespaces ...');

    const req: GetRegistryNamespacesData = {
      query: {
        limit: filterSortPagSearchState.pagination.limit,
        page: filterSortPagSearchState.pagination.page,
        sort_by: filterSortPagSearchState.sort?.key as 'name' | 'created_at',
        order: filterSortPagSearchState.sort?.order == -1 ? 'desc' : 'asc',
        search: filterSortPagSearchState.search_value,
      },
      url: '/registry/namespaces',
    };

    if (filterSortPagSearchState.filters && filterSortPagSearchState.filters.length != 0) {
      filterSortPagSearchState.filters.forEach((f) => {
        if (f.key === 'state' && req.query && f.values[0]) {
          // TODO: Current filter is one of states. But need to support array of values
          // Need to check what backend supports now
          req.query.state = f.values[0] as 'Active' | 'Deprecated' | 'Disabled';
        }

        if (f.key === 'purpose' && req.query && f.values.length != 0) {
          req.query.purpose = f.values[0] as 'Team' | 'Project';
        }

        //TODO: add another filter key visibility
      });
    }

    const { data, error } = await getRegistryNamespaces(req);
    hideLoading();

    if (error) {
      showError(error.error_message);
    } else {
      setTotalEntries(data.total);
      if (data.entities) {
        for (let i = 0; i < data.total - data.entities.length; i++) {
          data.entities.push({} as NamespaceViewDto); // adding fake data for showing pagination properly
        }
        setTableData(data.entities);
      }
    }
  }, [
    filterSortPagSearchState.filters,
    filterSortPagSearchState.pagination.limit,
    filterSortPagSearchState.pagination.page,
    filterSortPagSearchState.search_value,
    filterSortPagSearchState.sort?.key,
    filterSortPagSearchState.sort?.order,
    hideLoading,
    showError,
    showLoading,
  ]);

  const handleManageAccess = (id: string) => {
    navigate('/console/management/namespaces/' + id);
  };

  const handleFilter = (options: string[]) => {
    const statuses: string[] = [];
    const purpose: string = '';
    const isPublic: boolean = false;
  };

  const handleSearch = (searchTerm: string): void => { };

  return (
    <div className="flex-grow-1 flex flex-column gap-3 ">
      <div className="border-round-lg  flex flex-column gap-2">
        <div className="flex flex-column my-2 gap-2">
          <div className="flex flex-row align-items-stretch justify-content-between max-h-4rem h-3rem">
            <div className="pl-2 text-2xl title-500">Namespaces</div>
            <div className="flex-grow-1"></div>
            <div className="pr-2 flex flex-row pt-1 pb-1 justify-content-end gap-4  h-100">
              <SearchButton placeholder="Search namespaces ...." handleSearch={handleSearch} />
              <Button
                className="border-1 pl-3 pr-3  border-solid border-round-3xl border-teal-100 text-xs shadow-1"
                onClick={() => setShowCreateNamespaceDialog((c) => !c)}
              >
                <span className="pi pi-plus text-sm "></span>
                &nbsp;&nbsp;&nbsp;&nbsp;
                <span className="font-inter text-sm  font-semibold">Create Namespace</span>
              </Button>
            </div>
          </div>
          <div className="flex flex-row align-items-center pl-3 pr-2  mt-2 shadow-1 bg-white border-round-xl ">
            <div className=" text-sm"> Select Filter</div>
            <ChipsFilter
              filterOptions={NAMESPACE_FILTER_OPTIONS}
              handleFilterChange={handleFilter}
            />
          </div>
        </div>
        <DataTable
          value={talbeData}
          paginator
          rows={8}
          totalRecords={totalEntries}
          pt={{
            wrapper: {
              className: 'bg-white shadow-1 border-round-top-2xl',
            },
            root: {
              className: 'bg-white shadow-1 border-round-2xl',
            },
            footer: {
              className: 'bg-white shadow-1 border-round-2xl',
            },
            paginator: {
              root: {
                className: 'bg-white border-round-bottom-2xl pb-0 mb-0',
              },
            },
          }}
          className="bg-white shadow-1 border-round-sm"
          paginatorLeft={<div className="text-xs">{15} namespaces found</div>}
        // onPage={handlePagination}
        >
          <Column header="Name" body={NameColumnTemplate} sortable />
          <Column header="Type" field="type" />
          <Column
            header="Maintainers"
            body={(ns: NamespaceInfo) => {
              return (
                <div className="flex flex-row align-items-end gap-1">
                  <span>john, supun,</span>
                  <span className="pi pi-ellipsis-h cursor-pointer text-blue-600 text-xs"></span>
                  <span className="pi pi-external-link cursor-pointer text-blue-600 text-xs"></span>
                </div>
              );
            }}
          />

          <Column
            header="Developers"
            body={(ns: NamespaceInfo) => {
              return (
                <div className="flex flex-row align-items-end gap-1">
                  <span>john, supun,</span>
                  <span className="pi pi-ellipsis-h cursor-pointer text-blue-600 text-xs"></span>
                  <span className="pi pi-external-link cursor-pointer text-blue-600 text-xs"></span>
                </div>
              );
            }}
          />

          {/* <Column header="Repository" body={RepoColumnTemplate} /> */}
          {/* <Column header="Access" body={AccessColumnTemplate} /> */}
          <Column
            header="Created At"
            sortable
            body={(ns: NamespaceInfo) => ns.created_at?.toUTCString()}
          />

          <Column
            body={(ns: NamespaceInfo) => (
              <Button
                className="border-round-3xl border-1 text-sm"
                outlined
                size="small"
                onClick={() => handleManageAccess(ns.id)}
              >
                <span className="pi pi-eye text-sm"></span>
                &nbsp;&nbsp;
                <span className="text-sm">View</span>
              </Button>
            )}
          />
        </DataTable>
        <CreateNamespaceDialog
          visible={showCreateNamespaceDialog}
          hideCallback={() => setShowCreateNamespaceDialog(false)}
        />
      </div>
    </div>
  );
};

export default NamespaceMgmtPage;
