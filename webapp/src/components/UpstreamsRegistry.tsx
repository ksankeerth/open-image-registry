import { Badge } from "primereact/badge";
import { Button } from "primereact/button";
import { Column } from "primereact/column";
import { DataTable } from "primereact/datatable";
import { Divider } from "primereact/divider";
import { InputText } from "primereact/inputtext";
import { Sidebar } from "primereact/sidebar";
import React, { useEffect, useState, useRef } from "react";
import { PostUpstreamRequestBody } from "../types/request_response";
import { Checkbox } from "primereact/checkbox";

const UpstreamRegistry = (props: {
  visible: boolean;
  hideCallback: React.Dispatch<React.SetStateAction<boolean>>;
}) => {
  const [showFormToAddUpstream, setShowFormToAddUpstream] =
    useState<boolean>(true);

  const [isFormStep1Filled, setFormStep1Filled] = useState<boolean>(false);
  const [isFormStep2Filled, setFormStep2Filled] = useState<boolean>(false);

  const [tabIndexOfAddForm, setTabIndexOfAddForm] = useState<1 | 2>(1);

  const [newUpstreamRequest, setNewUpstreamRequest] =
    useState<PostUpstreamRequestBody>();

  useEffect(() => {
    if (
      newUpstreamRequest?.registry_name != "" &&
      newUpstreamRequest?.registry_url != "" &&
      newUpstreamRequest?.auth?.username != "" &&
      newUpstreamRequest?.auth?.password != "" &&
      Number(newUpstreamRequest?.storage?.limit) > 0 &&
      Number(newUpstreamRequest?.storage?.cleanup_threshold) > 0 &&
      Number(newUpstreamRequest?.storage?.cleanup_threshold) < 100
    ) {
      setFormStep1Filled(true);
    }

    if (
      newUpstreamRequest?.cache?.offline_mode !== undefined &&
      Number(newUpstreamRequest?.cache.ttl) > 0 &&
      newUpstreamRequest?.proxy?.retries > 0 &&
      newUpstreamRequest?.proxy?.socket_timeout > 0 &&
      newUpstreamRequest?.proxy?.enable !== undefined
    ) {
      setFormStep2Filled(true);
    }
  }, [newUpstreamRequest]);

  return (
    <Sidebar
      position="bottom"
      visible={props.visible}
      onHide={() => props.hideCallback(false)}
      header={<div className="text-color text-lg">Upstream Registeries</div>}
      style={{ height: "70vh" }}
      className="flex align-items-stretch"
    >
      <div className="h-full flex-grow-1 flex gap-3">
        <div>
          <DataTable
            value={[
              {
                repository_name: "proxy_docker_registry",
                cached_image_count: 23,
              },
              {
                repository_name: "proxy_quay_registry",
                cached_image_count: 11,
              },
            ]}
          >
            <Column
              header={
                <div className="w-6rem">
                  {!showFormToAddUpstream && (
                    <Button size="small">Add New</Button>
                  )}
                </div>
              }
              body={<i className="pi pi-sync cursor-pointer text-blue-600"></i>}
            ></Column>
            <Column field="repository_name" header="Repository" />
            <Column field="cached_image_count" header="Cached Images" />
          </DataTable>
        </div>

        <Divider layout="vertical" className="h-full" />

        {/* <Fieldset legend="proxy_docker_registry" className="w-full">

      </Fieldset> */}

        {/* Form to create new repository */}
        <div className="flex flex-column gap-2 w-full">
          <div className="text-color text-lg">New Upstream Registry</div>
          <div className="flex flex-row mt-2">
            <Divider className="flex-grow-1">
              <Badge
                value={"1"}
                className={isFormStep1Filled ? "mr-2" : "mr-2 p-badge-outlined"}
              >
                1
              </Badge>
              <span className="text-sm">&nbsp;&nbsp;Basic Info</span>
            </Divider>
            <Divider className="flex-grow-1">
              <Badge
                value={"2"}
                className={
                  isFormStep1Filled && isFormStep2Filled
                    ? "mr-2"
                    : "mr-2 p-badge-outlined"
                }
              ></Badge>
              <span className="text-sm">&nbsp;&nbsp;Advance Config</span>
            </Divider>
            <Divider align="right">
              <Button
                size="small"
                disabled={!(isFormStep1Filled && isFormStep2Filled)}
              >
                Complete
              </Button>
            </Divider>
          </div>

          <div className="grid ml-2 mr-2 mt-1">
            {/* Basic Info */}
            {tabIndexOfAddForm == 1 && (
              <React.Fragment>
                {/* repository_key */}
                <div className="col-3">
                  <label
                    htmlFor="registry_name"
                    className="text-color-secondary font-medium text-sm"
                  >
                    Registry Name
                  </label>
                </div>
                <div className="col-9">
                  <InputText
                    id="registry_name"
                    aria-describedby="registry_name_help"
                    size={60}
                    className="border-1 text-xs"
                    value={newUpstreamRequest?.registry_name}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          registry_name: e.target.value,
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                </div>
                {/* URL */}
                <div className="col-3">
                  <label
                    htmlFor="registry_url"
                    className="text-color font-medium text-sm"
                  >
                    URL
                  </label>
                </div>
                <div className="col-9">
                  <InputText
                    id="registry_url"
                    aria-describedby="registry_url_help"
                    size={60}
                    className="border-1 text-xs"
                    value={newUpstreamRequest?.registry_url}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          registry_url: e.target.value,
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                </div>

                <div className="col-12 text-color font-medium">
                  Authentication
                </div>

                <div className="col-2">
                  <label
                    htmlFor="username"
                    className="text-color font-medium text-sm"
                  >
                    Username
                  </label>
                </div>
                <div className="col-4 flex text-xs align-items-center">
                  <InputText
                    id="username"
                    aria-describedby="username_help"
                    size={60}
                    className="border-1 text-xs"
                    value={newUpstreamRequest?.auth?.username}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          auth: {
                            ...current?.auth,
                            username: e.target.value,
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                </div>

                <div className="col-2">
                  <label
                    htmlFor="password"
                    className="text-color font-medium text-sm"
                  >
                    Password
                  </label>
                </div>
                <div className="col-4 flex text-xs align-items-center">
                  <InputText
                    id="password"
                    aria-describedby="password_help"
                    size={60}
                    className="border-1 text-xs"
                    type="password"
                    value={newUpstreamRequest?.auth?.password}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          auth: {
                            ...current?.auth,
                            password: e.target.value,
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                </div>

                <div className="col-12 text-color font-medium">Storage</div>

                <div className="col-2">
                  <label
                    htmlFor="Limit"
                    className="text-color font-medium text-sm"
                  >
                    Limit
                  </label>
                </div>
                <div className="col-3 flex text-xs align-items-center">
                  <InputText
                    id="limit"
                    aria-describedby="limit_help"
                    size={10}
                    className="border-1 text-xs"
                    type="number"
                    value={newUpstreamRequest?.storage?.limit?.toString()}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          storage: {
                            ...current?.storage,
                            limit: Number(e.target.value),
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <div className="pl-2 flex text-center">Gbs</div>
                </div>
                <div className="col-3">
                  <label
                    htmlFor="cleanup_threshold"
                    className="text-color font-medium text-sm"
                  >
                    Cleanup Threshold
                  </label>
                </div>
                <div className="col-3 flex text-xs align-items-center">
                  <InputText
                    id="cleanup_threshold"
                    aria-describedby="cleanup_threshold_help"
                    size={10}
                    className="border-1 text-xs"
                    type="number"
                    value={newUpstreamRequest?.storage?.cleanup_threshold?.toString()}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          storage: {
                            ...current?.storage,
                            cleanup_threshold: Number(e.target.value),
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <div className="w-3rem pl-2"> %</div>
                </div>
              </React.Fragment>
            )}

            {/* Advance Info */}
            {tabIndexOfAddForm == 2 && (
              <React.Fragment>
                <div className="col-12 text-color font-medium">Cache</div>
                <div className="col-3">
                  <label
                    htmlFor="cleanup_threshold"
                    className="text-color font-medium text-sm"
                  >
                    TTL
                  </label>
                </div>

                <div className="col-3 flex text-xs align-items-center">
                  <InputText
                    id="cleanup_threshold"
                    aria-describedby="cleanup_threshold_help"
                    size={10}
                    className="border-1 text-xs"
                    type="number"
                    value={newUpstreamRequest?.cache?.ttl?.toString()}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          cache: {
                            ...current?.cache,
                            ttl: Number(e.target.value),
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <div className="pl-2">Seconds</div>
                </div>

                <div className="col-6"></div>

                <div className="col-6">
                  <Checkbox
                    inputId="offline_mode"
                    name="pi"
                    value="Offline Mode for non-latest tags"
                    checked={Boolean(newUpstreamRequest?.cache?.offline_mode)}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          cache: {
                            ...current?.cache,
                            offline_mode: !current?.cache?.offline_mode,
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <label htmlFor="offline_mode" className="ml-2 text-sm">
                    Prevent upstream checks for non-latest tags
                  </label>
                </div>

                <div className="col-12 text-color font-medium">Proxy</div>
                <div className="col-3">
                  <Checkbox
                    inputId="proxy_enable"
                    name="pi"
                    value="Offline Mode for non-latest tags"
                    checked={Boolean(newUpstreamRequest?.proxy?.enable)}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          proxy: {
                            ...current?.proxy,
                            enable: current?.proxy.enable,
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <label htmlFor="proxy_enable" className="ml-2 text-sm">
                    Enable
                  </label>
                </div>
                <div className="col-2">
                  <label
                    htmlFor="proxy_url"
                    className="text-color font-medium text-sm"
                  >
                    URL
                  </label>
                </div>
                <div className="col-7">
                  <InputText
                    disabled={!Boolean(newUpstreamRequest?.proxy?.enable)}
                    id="proxy_url"
                    aria-describedby="proxy_url_help"
                    size={60}
                    className="border-1 text-xs"
                    value={newUpstreamRequest?.proxy?.url}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          proxy: {
                            ...current?.proxy,
                            url: e.target.value,
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                </div>

                <div className="col-3">
                  <label
                    htmlFor="socket_timeout"
                    className="text-color font-medium text-sm"
                  >
                    Socket Timeout
                  </label>
                </div>
                <div className="col-3 flex text-xs align-items-center">
                  <InputText
                    disabled={!Boolean(newUpstreamRequest?.proxy?.enable)}
                    id="proxy_socket_timeout"
                    aria-describedby="socket_timeout_help"
                    size={10}
                    className="border-1 text-xs"
                    type="number"
                    value={newUpstreamRequest?.proxy?.socket_timeout?.toString()}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          proxy: {
                            ...current?.proxy,
                            socket_timeout: Number(e.target.value),
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <div className="pl-2 flex text-center">Milliseconds</div>
                </div>
                <div className="col-3">
                  <label
                    htmlFor="retries_count"
                    className="text-color font-medium text-sm"
                  >
                    Retries
                  </label>
                </div>
                <div className="col-3 flex text-xs align-items-center">
                  <InputText
                    disabled={!Boolean(newUpstreamRequest?.proxy?.enable)}
                    id="retries_count"
                    aria-describedby="retries_count_help"
                    size={10}
                    className="border-1 text-xs"
                    type="number"
                    value={newUpstreamRequest?.proxy?.retries?.toString()}
                    onChange={(e) =>
                      setNewUpstreamRequest((current) => {
                        return {
                          ...current,
                          proxy: {
                            ...current?.proxy,
                            retries: Number(e.target.value),
                          },
                        } as PostUpstreamRequestBody;
                      })
                    }
                  />
                  <div className="w-3rem pl-2"> %</div>
                </div>
              </React.Fragment>
            )}

            <div className="col-12 flex justify-content-end">
              {tabIndexOfAddForm == 1 && (
                <Button size="small" onClick={() => setTabIndexOfAddForm(2)}>
                  Next
                </Button>
              )}
              {tabIndexOfAddForm == 2 && (
                <Button size="small" onClick={() => setTabIndexOfAddForm(1)}>
                  Back
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </Sidebar>
  );
};

export default UpstreamRegistry;
