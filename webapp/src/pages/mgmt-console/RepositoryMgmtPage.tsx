import { AutoComplete } from 'primereact/autocomplete';
import { Button } from 'primereact/button';
import { Divider } from 'primereact/divider';
import React from 'react';

const RepositoryAccessPage = () => {
  return (
    <div className="h-full  flex flex-column gap-2 p-2 pt-4">
      <div className="flex-grow-1 border-round-lg bg-white flex flex-column pt-2 gap-2 ">
        <div className="pl-3 font-semibold text-lg">Serving 10,243 Repositories!</div>
        <div className="flex-grow-1 flex flex-column align-items-center justify-content-center gap-3">
          <div className="font-semibold pb-3">Select Namespace and Repository to view</div>
          <div className="flex gap-4">
            <AutoComplete
              dropdown
              placeholder="Select Namespace"
              size={30}
              pt={{
                input: {
                  root: {
                    className: 'border-1',
                  },
                },
              }}
              className="border-none border-200 border-round-lg"
            />
            <AutoComplete
              dropdown
              placeholder="Select Repository"
              size={30}
              className="border-none border-200 border-round-lg"
              pt={{
                input: {
                  root: {
                    className: 'border-1',
                  },
                },
              }}
            />
            <Button size="small">View</Button>
          </div>
        </div>
        <div className="flex-grow-1 flex flex-column">
          <Divider layout="horizontal" align="center">
            <div className="font-medium">Recently Visited Repositories</div>
          </Divider>
          <div className="flex-grow-1 pb-3 flex flex-wrap align-items-end justify-content-between gap-3 pl-3 pr-3">
            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm">34</div>
              </div>
            </div>

            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 flex justify-content-end">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm flex justify-content-end">34</div>
              </div>
            </div>

            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 flex justify-content-end">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm flex justify-content-end">34</div>
              </div>
            </div>

            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 flex justify-content-end">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm flex justify-content-end">34</div>
              </div>
            </div>

            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 flex justify-content-end">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm flex justify-content-end">34</div>
              </div>
            </div>

            <div className="flex flex-column gap-3 border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 p-2">
              <div className="text-color font-medium flex gap-4 align-items-center justify-content-between">
                <div>
                  <span className="text-blue-600 cursor-pointer">team-platform-eng</span>/
                  <span className="text-blue-600 cursor-pointer">order-backend</span>
                </div>
              </div>
              <div className="flex justify-content-between">
                <div>
                  <span className="pi pi-tags text-sm"></span>&nbsp;&nbsp;<span>54</span>
                </div>
                <div>
                  <span className="pi pi-clock text-sm"></span> &nbsp;&nbsp; 2h ago
                </div>
              </div>
              <div className="grid grid-nogutter pt-2">
                <div className="col-5">Developers</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 flex justify-content-end">Guests</div>
                <div className="col-5 text-sm">23</div>
                <div className="col-2">
                  <Divider className="p-0 m-0" layout="vertical" />
                </div>
                <div className="col-5 text-sm flex justify-content-end">34</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RepositoryAccessPage;
