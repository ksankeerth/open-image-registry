import React, { useCallback, useEffect, useRef, useState } from 'react';

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
import { getResourceNamespaces, GetResourceNamespacesData, NamespaceViewDto } from '../../../api';
import { useLoader } from '../../../components/loader';
import { useToast } from '../../../components/ToastComponent';

const NameColumnTemplate = (ns: NamespaceViewDto) => {
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

    const req: GetResourceNamespacesData = {
      query: {
        limit: filterSortPagSearchState.pagination.limit,
        page: filterSortPagSearchState.pagination.page,
        sort_by: filterSortPagSearchState.sort?.key as 'name' | 'created_at',
        order: filterSortPagSearchState.sort?.order == -1 ? 'desc' : 'asc',
        search: filterSortPagSearchState.search_value,
      },
      url: '/resource/namespaces',
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

        if (f.key === 'is_public' && req.query && f.values.length != 0) {
          req.query.is_public = f.values[0] as boolean;
        }
      });
    }

    const { data, error } = await getResourceNamespaces(req);
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

  useEffect(() => {
    // TODO: Need to check this warning and why
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadNamespaces();
  }, [loadNamespaces]);

  const handleManageAccess = (id: string) => {
    navigate('/console/management/namespaces/' + id);
  };

  const handleFilter = (options: string[]) => {
    // current backend only support only one value. so we should not take multiple values
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
          paginatorLeft={<div className="text-xs">{totalEntries} namespaces found</div>}
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
          <Column header="Created At" sortable body={(ns: NamespaceViewDto) => ns.created_at} />

          <Column
            body={(ns: NamespaceViewDto) => (
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
