import React, { useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

import { Divider } from 'primereact/divider';
import UpstreamRegistry from './UpstreamsRegistry';
import ImagesView from './ImagesView';
import LogoComponent from './LogoComponent';
import { ToastProvider } from './ToastComponent';
import { postAuthLogout } from '../api';
import { MGMT_CONSOLE_BASE_PATH, REG_CONSOLE_BASE_PATH } from '../config/menus';
import { Tooltip } from 'primereact/tooltip';

const RootLayout = () => {
  const [showUpstreamModal, setShowUpstreamModal] = useState<boolean>(false);
  const [showImagesModal, setShowImagesModal] = useState<boolean>(false);

  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = async () => {
    const { data, error } = await postAuthLogout({});
    if (error) {
      console.log(error);
    }
    navigate('/login');
  };

  const navigateToManagementConsole = () => {
    navigate('/console/management', {
      viewTransition: true,
    });
  };

  const title = (): string => {
    const currentPath = location.pathname;
    if (currentPath.includes(REG_CONSOLE_BASE_PATH)) {
      return 'REGISTRY CONSOLE';
    } else if (currentPath.includes(MGMT_CONSOLE_BASE_PATH)) {
      return 'MANAGEMENT CONSOLE';
    }
    return '';
  };

  const isManagementConsoleView = (): boolean => {
    return location.pathname.includes(MGMT_CONSOLE_BASE_PATH);
  };

  const isRegistryConsoleView = (): boolean => {
    return location.pathname.includes(REG_CONSOLE_BASE_PATH);
  };

  const [availableHeight, setAvailableHeight] = useState(0);

  useEffect(() => {
    const calculate = () => {
      const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize);

      const headerHeightPx = 5 * rootFontSize;

      setAvailableHeight(window.innerHeight - headerHeightPx);
    };

    calculate();
    window.addEventListener('resize', calculate);
    return () => window.removeEventListener('resize', calculate);
  }, []);

  return (
    <ToastProvider>
      <div className="flex flex-column min-h-screen max-h-screen">
        <div className="flex-grow-0 pl-2 pr-2 h-5rem w-screen flex flex-row  justify-content-between mr-4 gap-0">
          <div className="pt-1 pb-1">
            <LogoComponent showNameInOneLine={false} />
          </div>

          {/* Title  */}
          <div className="flex-grow-1 flex align-items-end  pb-3">
            <div
              className=" flex-grow-1 flex flex-row align-items-end justify-content-center  text-color-secondary"
              style={{ fontFamily: 'Montserrat, sans-serif', fontWeight: 500 }}
            >
              {title()}
            </div>
          </div>

          <div className="flex-grow-o flex justify-content-end pr-2 align-items-center">
            <Tooltip
              target=".header-nav-icon"
              position="bottom"
              showDelay={300}
              pt={{
                text: {
                  className: 'text-gray-200 text-xs font-normal px-3 py-2 rounded-lg',
                  style: {
                    backgroundColor: '#374151',
                    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
                  },
                },
                arrow: {
                  style: {
                    borderBottomColor: '#374151',
                  },
                },
              }}
            />
            {!isRegistryConsoleView() && (
              <div
                className="header-nav-icon"
                data-pr-tooltip="Registry Console"
                data-pr-position="top"
                data-pr-at="center top-10"
              >
                <span
                  className="pi pi-objects-column text-teal-700 cursor-pointer"
                  style={{ fontSize: '1.2rem' }}
                ></span>
              </div>
            )}
            <Divider layout="vertical" className=" pb-1" />
            <div
              className="flex gap-4"
              style={{
                zIndex: 20,
              }}
            >
              <div
                className="header-nav-icon"
                data-pr-tooltip="Profile"
                data-pr-position="top"
                data-pr-at="center top-10"
              >
                <span
                  className="pi pi-user text-teal-700 cursor-pointer"
                  style={{ fontSize: '1.2rem' }}
                ></span>
              </div>

              {!isManagementConsoleView() && (
                <div
                  className="header-nav-icon cursor-pointer"
                  data-pr-tooltip="Management Console"
                  data-pr-position="top"
                  data-pr-at="center top-10"
                  onClick={navigateToManagementConsole}
                >
                  <span
                    className="pi pi-cog text-teal-700 cursor-pointer"
                    style={{ fontSize: '1.2rem' }}
                  ></span>
                </div>
              )}

              <div
                className="header-nav-icon cursor-pointer"
                data-pr-tooltip="Logout"
                data-pr-position="top"
                data-pr-at="left-10 top-10"
                onClick={handleLogout}
              >
                <span
                  className="pi pi-sign-out text-teal-700 cursor-pointer"
                  style={{ fontSize: '1.2rem' }}
                ></span>
              </div>
            </div>
          </div>
        </div>
        <Divider layout="horizontal" className="p-0 m-0" />
        <div className="flex-grow-1 flex align-items-stretch">
          <Outlet context={{ availableHeight: availableHeight }} />
        </div>
        {/* For upstreams */}
        <UpstreamRegistry visible={showUpstreamModal} hideCallback={setShowUpstreamModal} />

        {/* For Images view */}
        <ImagesView visible={showImagesModal} hideCallback={setShowImagesModal} />
      </div>
    </ToastProvider>
  );
};

export default RootLayout;
