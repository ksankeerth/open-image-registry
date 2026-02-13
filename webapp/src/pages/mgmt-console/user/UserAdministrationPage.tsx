import React, { useCallback, useEffect, useRef, useState } from 'react';

import { DataTable, DataTableStateEvent } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { Button } from 'primereact/button';
import { OverlayPanel } from 'primereact/overlaypanel';
import { UserAccountViewDto, GetUsersData, getUsers } from '../../../api';
import ChipsFilter from '../../../components/ChipsFilter';
import CreateUserAccountDialog from '../../../components/CreateUserAccount';
import { useLoader } from '../../../components/loader';
import SearchButton from '../../../components/SearchButton';
import { useToast } from '../../../components/ToastComponent';
import UserAccountView from '../../../components/UserAccountView';
import { USERS_FILTER_OPTIONS } from '../../../config/table_filter';
import {
  TableFilterSearchPaginationSortState,
  TableColumnFilterState,
} from '../../../types/app_types';

const AccountStatusBodyTemplate = (user: UserAccountViewDto) => {
  const lockOpref = useRef<OverlayPanel>(null);
  const incompleteOpRef = useRef<OverlayPanel>(null);
  const activeOpRef = useRef<OverlayPanel>(null);

  return (
    <React.Fragment>
      {user.locked && (
        <Button
          className="p-1 bg-transparent border-none hover:bg-gray-100"
          onClick={(e) => {
            lockOpref.current?.toggle(e);
          }}
        >
          <span className="pi pi-lock text-sm text-red-500 cursor-pointer"></span>
          <OverlayPanel ref={lockOpref}>
            <div className="bg-white  flex flex-column gap-2 text-xs">
              <div className="font-semibold">User Account was locked for:</div>
              <div className="">{user.locked_reason}</div>
              {!user.locked_reason?.includes('New') && (
                <div className="flex w-full justify-content-end">
                  <Button size="small" className="text-xs p-1 m-0">
                    Unlock
                  </Button>
                </div>
              )}

              {!user.password_recovery_reason?.includes('New') && (
                <React.Fragment>
                  <div className="font-semibold">Password Recovery is in progress:</div>
                  <div>{user.password_recovery_reason}</div>
                </React.Fragment>
              )}
            </div>
          </OverlayPanel>
        </Button>
      )}
      {!user?.locked && !user?.password_recovery_id && (
        <Button
          className="p-1 bg-transparent border-none hover:bg-gray-100"
          onClick={(e) => {
            activeOpRef.current?.toggle(e);
          }}
        >
          <span className="pi pi-verified text-sm text-teal-500 cursor-pointer"></span>
          <OverlayPanel ref={activeOpRef}>
            <div className="bg-white  flex flex-column gap-2 text-xs">
              <div className="font-semibold">Verified User account. Created at</div>
              <div className="">{user.created_at}</div>
              <div className="flex w-full justify-content-end">
                <Button size="small" severity="danger" className="text-xs p-1 m-0">
                  Lock
                </Button>
              </div>
            </div>
          </OverlayPanel>
        </Button>
      )}
      {!user?.locked && user?.password_recovery_id && (
        <Button
          className="p-1 bg-transparent border-none hover:bg-gray-100"
          onClick={(e) => {
            incompleteOpRef.current?.toggle(e);
          }}
        >
          <span className="pi pi-exclamation-circle text-yellow-500 cursor-pointer"></span>
          <OverlayPanel ref={incompleteOpRef}>
            <div className="bg-white  flex flex-column gap-2 text-xs">
              <div className="font-semibold">Password Recovery is in-progress:</div>
              <div className="">{user.password_recovery_reason}</div>
              <div className="flex w-full justify-content-end">
                <Button size="small" severity="danger" className="text-xs p-1 m-0">
                  Lock
                </Button>
              </div>
            </div>
          </OverlayPanel>
        </Button>
      )}
    </React.Fragment>
  );
};

const UsernameColumnBodyTemplate = (
  user: UserAccountViewDto,
  hideCallBack: (reloadUsers: boolean) => void
) => {
  const [showUserAccountView, setShowUserAccountView] = useState<boolean>(false);
  return (
    <div>
      <Button
        className="p-1 bg-transparent text-blue-600 border-none hover:border-none hover:bg-gray-100 hover:underline"
        onClick={() => setShowUserAccountView((c) => !c)}
      >
        {user.username}
      </Button>
      <UserAccountView
        visible={showUserAccountView}
        account={user}
        hideCallback={() => {
          setShowUserAccountView((c) => !c);
          hideCallBack(true);
        }}
      />
    </div>
  );
};

const UserAdministrationPage = () => {
  const [showUserAccountView, setShowUserAccountView] = useState<boolean>(false);

  const [showCreateUserAccountDialog, setShowCreateUserAccountDialog] = useState<boolean>(false);

  const [filterSortPagSearchState, setFilterSortPagSearchState] = useState<
    TableFilterSearchPaginationSortState<UserAccountViewDto>
  >({
    pagination: {
      page: 1,
      limit: 8,
    },
    sort: {
      key: 'username',
      order: 1,
    },
  });

  const [talbeData, setTableData] = useState<UserAccountViewDto[]>([]);
  const [totalEntries, setTotalEntries] = useState<number>(0);

  const { showError } = useToast();

  const { showLoading, hideLoading } = useLoader();

  const loadUsers = useCallback(async () => {
    showLoading('Loading users ...');

    const req: GetUsersData = {
      query: {
        limit: filterSortPagSearchState.pagination.limit,
        page: filterSortPagSearchState.pagination.page,
        sort_by: filterSortPagSearchState.sort?.key as
          | 'username'
          | 'email'
          | 'role'
          | 'display_name'
          | 'last_loggedin_at',
        order: filterSortPagSearchState.sort?.order == -1 ? 'desc' : 'asc',
        search: filterSortPagSearchState.search_value,
      },
      url: '/users',
    };

    if (filterSortPagSearchState.filters && filterSortPagSearchState.filters.length != 0) {
      filterSortPagSearchState.filters.forEach((f) => {
        if (f.key === 'locked' && req.query && f.values[0]) {
          req.query.locked = f.values[0] as boolean;
        }

        if (f.key === 'role' && req.query && f.values.length != 0) {
          req.query.role = f.values as Array<'Admin' | 'Developer' | 'Maintainer' | 'Guest'>;
        }
      });
    }

    const { data, error } = await getUsers(req);

    hideLoading();

    if (error) {
      showError(error.error_message);
    } else {
      setTotalEntries(data.total);
      if (data.entities) {
        for (let i = 0; i < data.total - data.entities.length; i++) {
          data.entities.push({} as UserAccountViewDto); // adding fake data for showing pagination properly
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
    //TODO: We have to revisit this warning
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadUsers();
  }, [loadUsers]);

  const handleDialogHide = (reloadUsers: boolean) => {
    if (reloadUsers) {
      loadUsers();
    }

    setShowCreateUserAccountDialog(false);
    setShowUserAccountView(false);
  };

  const handleFilter = (options: string[]) => {
    const roles: string[] = [];
    const statuses: string[] = [];

    const filters: TableColumnFilterState<UserAccountViewDto>[] = [];

    options.forEach((v) => {
      if (v.startsWith('Role:')) {
        roles.push(v.slice('Role:'.length));
      }
      if (v.includes('Status:')) {
        statuses.push(v.slice('Status:'.length));
      }
    });
    if (roles.length < USERS_FILTER_OPTIONS.map((v) => v.value.startsWith('Role:')).length) {
      const roleFilter = {
        key: 'role' as keyof UserAccountViewDto,
        values: roles,
      };
      filters.push(roleFilter);
    }
    if (statuses.length == 1) {
      const statusFilter = {
        key: 'locked' as keyof UserAccountViewDto,
        values: [statuses[0] == 'locked' ? true : false],
      };
      filters.push(statusFilter);
    }
    setFilterSortPagSearchState((s: TableFilterSearchPaginationSortState<UserAccountViewDto>) => {
      return {
        ...s,
        filters: filters,
      };
    });
  };

  const handleSort = (event: DataTableStateEvent) => {
    //event.sortOder always 1

    setFilterSortPagSearchState((s: TableFilterSearchPaginationSortState<UserAccountViewDto>) => {
      let newSortOrder: 1 | 0 | -1 = 0;
      if (s.sort?.key == event.sortField) {
        if (s.sort?.order == 1) {
          newSortOrder = -1;
        } else if (s.sort?.order == -1) {
          newSortOrder = 1;
        }
      } else {
        newSortOrder = 1;
      }

      const newSort = {
        key: event.sortField as keyof UserAccountViewDto,
        order: newSortOrder,
      };
      if (s) {
        return {
          ...s,
          sort: newSort,
        };
      }

      return {
        sort: newSort,
        pagination: {
          page: 1,
          limit: 8,
        },
      };
    });
  };

  const handlePagination = (event: DataTableStateEvent) => {
    // DataTableStateEvent.page starts from zero. But our filter's page starts from 1;
    setFilterSortPagSearchState((s: TableFilterSearchPaginationSortState<UserAccountViewDto>) => {
      return {
        ...s,
        pagination: {
          page: event.page ? event.page + 1 : 1,
          limit: event.pageCount ? event.pageCount : 8,
        },
      };
    });
  };

  const handleSearch = (searchTerm: string): void => {
    // adding a small delay
    setTimeout(() => {
      setFilterSortPagSearchState((s: TableFilterSearchPaginationSortState<UserAccountViewDto>) => {
        if (searchTerm) {
          return {
            ...s,
            search_value: searchTerm,
          };
        } else {
          return {
            ...s,
            search_value: undefined,
          };
        }
      });
    }, 100);
  };

  return (
    <div className="flex flex-column flex-grow-1 gap-3">
      <div className="flex flex-column gap-2">
        <div className="flex flex-column my-2 gap-2">
          <div className="flex flex-row align-items-stretch justify-content-between max-h-4rem h-3rem">
            <div className="text-2xl title-500 pl-2">Users</div>
            <div className="flex-grow-1"></div>
            <div className="pr-2 flex flex-row pt-1 pb-1 justify-content-end gap-4  h-100">
              <SearchButton placeholder="Search users ...." handleSearch={handleSearch} />
              <Button
                className="border-1 pl-3 pr-3  border-solid border-round-3xl border-teal-100 text-xs shadow-1"
                onClick={() => setShowCreateUserAccountDialog((c) => !c)}
              >
                <span className="pi pi-plus text-sm "></span>
                &nbsp;&nbsp;&nbsp;&nbsp;
                <span className="font-inter text-sm  font-semibold">Create User</span>
              </Button>
            </div>
          </div>
          <div className="flex flex-row align-items-center pl-3 pr-2  mt-2 shadow-1 bg-white border-round-xl">
            <div className=" text-sm"> Select Filter</div>
            <ChipsFilter filterOptions={USERS_FILTER_OPTIONS} handleFilterChange={handleFilter} />
          </div>
        </div>
        <DataTable
          value={talbeData}
          paginator
          rows={filterSortPagSearchState.pagination.limit}
          totalRecords={totalEntries}
          paginatorLeft={<div className="text-xs">{totalEntries} users found</div>}
          onPage={handlePagination}
          onSort={handleSort}
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
        >
          <Column body={AccountStatusBodyTemplate} />
          <Column
            field="username"
            header="Username"
            body={(data) =>
              UsernameColumnBodyTemplate(data as UserAccountViewDto, handleDialogHide)
            }
            sortable
          // sortFunction={handleSort}
          />

          <Column field="email" header="Email" sortable />
          <Column field="role" header="Role" sortable />
          <Column field="display_name" header="Display Name" sortable />
          <Column
            field="last_loggedin_at"
            header="Last Logged In"
            sortable
            body={(user: UserAccountViewDto) => {
              return user.last_loggedin_at;
            }}
          />
        </DataTable>
        <CreateUserAccountDialog
          visible={showCreateUserAccountDialog}
          hideCallback={handleDialogHide}
        />
      </div>
    </div>
  );
};

export default UserAdministrationPage;
