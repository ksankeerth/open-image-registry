import React, { useEffect, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import { Divider } from "primereact/divider";
import UpstreamRegistry from "./UpstreamsRegistry";
import ImagesView from "./ImagesView";
import LogoComponent from "./LogoComponent";
import { ToastProvider } from "./ToastComponent";
import LoginPage from "../pages/LoginPage";
import AdminPanelComponent from "../pages/console/RegistryConsolePage";
import NewAccountSetupPage from "../pages/NewAccountSetupPage";

const RootLayout = () => {
  const [showUpstreamModal, setShowUpstreamModal] = useState<boolean>(false);
  const [showImagesModal, setShowImagesModal] = useState<boolean>(false);
  const [showAdminPanel, setShowAdminPanel] = useState<boolean>(false);

  const navigate = useNavigate();

  // TODO: for the moment we'll use local storage and later we have to decie how to handle
  // authentication and session properly
  const [authenticated, setAuthenticated] = useState<boolean>(
    Boolean(localStorage.getItem("authenticated"))
  );

  useEffect(() => {
    if (!authenticated) {
      navigate("/login");
    }
  }, [authenticated])

  const handleLogout = () => {
    localStorage.removeItem("authenticated");
    navigate("/login");
  };

  const navigateToManagementConsole = () => {
    navigate("/console", {
      viewTransition: true
    })
  }



  return (
    <ToastProvider>
      {authenticated && (
        <div className="flex flex-column min-h-screen max-h-screen">
          <div className="flex-grow-0  h-5rem w-screen flex flex-row  align-items-center justify-content-start mr-4 gap-0">
            <div className="flex-grow-1 flex justify-content-start">
              <LogoComponent showNameInOneLine={false} />
            </div>

            <div className="flex-grow-0 flex justify-content-start align-items-center gap-3 text-teal-700">
              <div
                className="cursor-pointer "
                style={{ zIndex: 50 }}
                onClick={() => setShowImagesModal(true)}
              >
                Images
              </div>
              <div
                className="cursor-pointer"
                style={{ zIndex: 50 }}
                onClick={() => setShowUpstreamModal(true)}
              >
                Upstreams
              </div>
              {/* <div className="cursor-pointer" style={{ zIndex: 50 }}>
                Settings
              </div> */}
            </div>
            <div className="flex-grow-o flex justify-content-end pr-3 align-items-center">
              <Divider layout="vertical" className=" pb-1" />
              <div
                className="flex gap-5"
                style={{
                  zIndex: 20,
                }}
              >
                <span
                  className="pi pi-user text-teal-700 cursor-pointer"
                  style={{ fontSize: "1.2rem" }}
                ></span>
                <span
                  className="pi pi-objects-column text-teal-700 cursor-pointer"
                  style={{ fontSize: "1.2rem" }}
                  onClick={navigateToManagementConsole}
                ></span>
                <span
                  className="pi pi-sign-out text-teal-700 cursor-pointer"
                  style={{ fontSize: "1.2rem" }}
                  onClick={handleLogout}
                ></span>
              </div>
            </div>
          </div>
          <div className="flex-grow-1 flex align-items-stretch">
            <Outlet />
          </div>
          {/* For upstreams */}
          <UpstreamRegistry
            visible={showUpstreamModal}
            hideCallback={setShowUpstreamModal}
          />

          {/* For Images view */}
          <ImagesView
            visible={showImagesModal}
            hideCallback={setShowImagesModal}
          />
        </div>
      )}
    </ToastProvider>
  );
};

export default RootLayout;