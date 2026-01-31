import React, { useEffect, useMemo, useState } from 'react';
import { Outlet, useLocation, useNavigate, useOutletContext } from 'react-router-dom';
import SideMenuList from './SideMenu';
import { BreadCrumb } from 'primereact/breadcrumb';
import { MenuEntity } from '../types/app_types';
import { MenuItem } from 'primereact/menuitem';

type ConsoleWrapperProps = {
  menus: MenuEntity[];
  menuKeyLinkMap: { key: string; nav_link: string }[];
  basePath: string;
  defaultPath: string;
};

const ConsoleWrapper = (props: ConsoleWrapperProps) => {
  const { availableHeight } = useOutletContext<{ availableHeight: number }>();

  const location = useLocation();
  const navigate = useNavigate();

  const [breadcrumpList, setBreadcrumpList] = useState<MenuItem[]>([
    { label: 'Home' },
    { label: 'Managment Console' },
    { label: 'Users' },
  ]);

  const selectedMenuKey = useMemo(() => {
    const currentPath = location.pathname;

    const matched = props.menuKeyLinkMap
      .filter((r) => currentPath.startsWith(r.nav_link))
      .sort((a, b) => b.nav_link.length - a.nav_link.length)[0];

    return matched ? matched.key : props.menus[0].children[0].key;
  }, [location.pathname, props.menuKeyLinkMap, props.menus]);

  useEffect(() => {
    if (location.pathname.startsWith('/console')) {
      if (location.pathname == props.basePath) {
        navigate(props.defaultPath, {
          replace: true,
          viewTransition: true,
        });
      }
    }
  }, [location.pathname, navigate, props.basePath, props.defaultPath]);

  const menuCollapsed = () => {
    // calculateHeight();
  };

  return (
    <div className="flex flex-column min-w-screen max-w-screen">
      <div className="flex flex-row min-h-full w-full">
        <SideMenuList
          menus={props.menus}
          selectedMenuKey={selectedMenuKey}
          menuCollapsed={menuCollapsed}
          availableHeight={availableHeight}
        />
        <div
          className="w-full  bg-offwhite px-4 flex flex-column flew-grow-1"
          style={{ height: availableHeight, overflowY: 'auto' }}
        >
          <div className="pt-4 pl-2">
            <BreadCrumb
              model={breadcrumpList}
              className="bg-offwhite border-none text-sm font-medium p-0 m-0"
              separatorIcon={<span className="pi pi-chevron-right text-xs" />}
            />
          </div>
          <div className="pt-2 pb-3">
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConsoleWrapper;